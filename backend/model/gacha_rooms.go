package model

import (
	"errors"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const GachaRoomTTL = 5 * time.Minute

var ErrGachaRoom = errors.New("room is full, closed, or outside the power range")

// Rooms share the arena gate, seats and escrow ledger with random matchmaking.
// A seat can belong to exactly one queue or room. Joining alone never charges.
type GachaArenaRoom struct {
	ID         string `gorm:"type:varchar(36);primaryKey" json:"id"`
	Status     string `gorm:"type:varchar(16);not null;index" json:"status"`
	StakeE     int    `gorm:"not null" json:"stake_e"`
	HostID     int    `gorm:"not null" json:"host_id"`
	GuestID    int    `json:"guest_id"`
	HostName   string `gorm:"type:text;not null" json:"host_name"`
	GuestName  string `gorm:"type:text" json:"guest_name"`
	HostPower  int    `gorm:"not null" json:"host_power"`
	GuestPower int    `json:"guest_power"`
	CreatedAt  int64  `json:"created_at"`
	ExpireAt   int64  `gorm:"index" json:"expire_at"`
	MatchID    string `gorm:"type:varchar(36)" json:"-"`
}

type GachaArenaRoomView struct {
	GachaArenaRoom
	HostReady  bool              `json:"host_ready"`
	GuestReady bool              `json:"guest_ready"`
	Match      *GachaArenaResult `json:"match,omitempty"`
}

type GachaArenaLobby struct {
	Rooms  []GachaArenaRoom    `json:"rooms"`
	Room   *GachaArenaRoomView `json:"room,omitempty"`
	Quota  int                 `json:"quota"`
	Queued bool                `json:"queued"`
}

func closeGachaRoom(tx *gorm.DB, id, status string, deltas map[int]int) error {
	var room GachaArenaRoom
	if err := tx.First(&room, "id = ?", id).Error; err != nil {
		return err
	}
	if room.Status != "open" {
		return nil
	}
	var seats []GachaArenaSeat
	if err := tx.Where("room_id = ?", id).Find(&seats).Error; err != nil {
		return err
	}
	for i := range seats {
		seat := &seats[i]
		if seat.Status == "waiting" {
			if err := refundGachaArenaSeat(tx, seat, status, deltas); err != nil {
				return err
			}
		} else if seat.Status == "room" {
			if err := tx.Model(seat).Updates(map[string]any{"status": status, "snapshot": ""}).Error; err != nil {
				return err
			}
		}
	}
	return tx.Model(&room).Update("status", status).Error
}

func gachaRoomView(tx *gorm.DB, userID int, room *GachaArenaRoom) (*GachaArenaRoomView, error) {
	if room.HostID != userID && room.GuestID != userID {
		return nil, ErrGachaIntent
	}
	view := &GachaArenaRoomView{GachaArenaRoom: *room}
	var seats []GachaArenaSeat
	if err := tx.Where("room_id = ? AND status IN ?", room.ID, []string{"waiting", "matched"}).Find(&seats).Error; err != nil {
		return nil, err
	}
	for _, seat := range seats {
		if seat.UserID == room.HostID {
			view.HostReady = true
		}
		if seat.UserID == room.GuestID {
			view.GuestReady = true
		}
	}
	if room.MatchID != "" {
		var err error
		view.Match, err = gachaArenaResult(tx, userID, room.MatchID)
		if err != nil {
			return nil, err
		}
	}
	return view, nil
}

func applyGachaRoomAction(userID int, intent GachaActionIntent, now time.Time) (*GachaActionResult, error) {
	intent.RequestID = strings.ToLower(intent.RequestID)
	hash, err := gachaIntentHash(intent)
	if err != nil {
		return nil, err
	}
	result := &GachaActionResult{}
	err = gachaArenaTransaction(func(tx *gorm.DB, deltas map[int]int) error {
		player, state, err := loadGachaPlayer(tx, userID, now)
		if err != nil {
			return err
		}
		if replay, err := gachaReplayResult(tx, userID, intent.RequestID, hash, result); replay || err != nil {
			return err
		}
		*result = GachaActionResult{}
		var seat GachaArenaSeat
		if err := tx.First(&seat, "user_id = ?", userID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var room GachaArenaRoom
		if intent.RoomID != nil {
			if err := tx.First(&room, "id = ?", strings.ToLower(*intent.RoomID)).Error; err != nil {
				return ErrGachaRoom
			}
		}
		if intent.Action == "room_leave" {
			if room.HostID != userID && room.GuestID != userID {
				return ErrGachaIntent
			}
			if err := closeGachaRoom(tx, room.ID, "cancelled", deltas); err != nil {
				return err
			}
		} else {
			if !GachaArenaEnabled() {
				return ErrGachaArenaPending
			}
			if state.Daily.ArenaPlayed >= GachaArenaDailyLimit {
				return ErrGachaArenaLimit
			}
			switch intent.Action {
			case "room_create", "room_join":
				if seat.Status == "waiting" || seat.Status == "room" {
					return ErrGachaArenaWaiting
				}
				power, err := GachaSquadPower(state)
				if err != nil || power <= 0 {
					return ErrGachaTeam
				}
				var user User
				if err := tx.Select("id", "username", "display_name", "status").First(&user, userID).Error; err != nil {
					return err
				}
				if user.Status != common.UserStatusEnabled {
					return ErrGachaIntent
				}
				name := user.DisplayName
				if name == "" {
					name = user.Username
				}
				if intent.Action == "room_create" {
					room = GachaArenaRoom{ID: uuid.NewString(), Status: "open", StakeE: *intent.StakeE, HostID: userID, HostName: name, HostPower: power, CreatedAt: now.UnixMilli(), ExpireAt: now.Add(GachaRoomTTL).UnixMilli()}
					if err := tx.Create(&room).Error; err != nil {
						return err
					}
				} else {
					minimum, maximum := gachaArenaPowerRange(room.HostPower)
					if room.Status != "open" || room.ExpireAt <= now.UnixMilli() || room.GuestID != 0 || room.HostID == userID || power < minimum || power > maximum {
						return ErrGachaRoom
					}
					var host User
					if err := tx.Select("id", "status").First(&host, room.HostID).Error; err != nil || host.Status != common.UserStatusEnabled {
						return ErrGachaRoom
					}
					room.GuestID, room.GuestName, room.GuestPower = userID, name, power
					if err := tx.Save(&room).Error; err != nil {
						return err
					}
				}
				snapshot, err := freezeGachaRoster(state)
				if err != nil {
					return err
				}
				seat = GachaArenaSeat{UserID: userID, ID: uuid.NewString(), Status: "room", RoomID: room.ID, StakeE: room.StakeE, Power: power, CreatedAt: now.UnixMilli(), ExpireAt: room.ExpireAt, Snapshot: GachaDocument(snapshot)}
				if err := tx.Save(&seat).Error; err != nil {
					return err
				}
			case "room_ready":
				if seat.RoomID != room.ID || (room.HostID != userID && room.GuestID != userID) {
					return ErrGachaIntent
				}
				if room.Status != "open" || room.ExpireAt <= now.UnixMilli() || room.GuestID == 0 {
					return ErrGachaRoom
				}
				if seat.Status != "room" && seat.Status != "waiting" {
					return ErrGachaRoom
				}
				otherID := room.HostID
				if userID == room.HostID {
					otherID = room.GuestID
				}
				var other GachaArenaSeat
				if err := tx.First(&other, "user_id = ? AND room_id = ?", otherID, room.ID).Error; err != nil {
					return ErrGachaRoom
				}
				if seat.Status == "room" {
					amount := room.StakeE * GachaArenaQuotaPerE
					r := tx.Model(&User{}).Where("id = ? AND status = ? AND quota >= ? AND quota <= ?", userID, common.UserStatusEnabled, amount, common.MaxWalletQuota-amount).Update("quota", gorm.Expr("quota - ?", amount))
					if r.Error != nil {
						return r.Error
					}
					if r.RowsAffected != 1 {
						return ErrGachaArenaFunds
					}
					deltas[userID] -= amount
					seat.Status = "waiting"
					if err := tx.Model(&seat).Update("status", seat.Status).Error; err != nil {
						return err
					}
					if err := tx.Create(&GachaArenaEntry{QueueID: seat.ID, Phase: "reserve", UserID: userID, DeltaQuota: -amount, Reason: "room", CreatedAt: now.UnixMilli()}).Error; err != nil {
						return err
					}
				}
				if other.Status == "waiting" {
					if err := settleGachaArenaPair(tx, state, &seat, &other, now, deltas); err != nil {
						return err
					}
					room.Status, room.MatchID = "matched", seat.MatchID
					if err := tx.Save(&room).Error; err != nil {
						return err
					}
				}
			}
		}
		if err := tx.First(&room, "id = ?", room.ID).Error; err != nil {
			return err
		}
		if err := tx.First(&seat, "user_id = ?", userID).Error; err != nil {
			return err
		}
		if err := storeGachaPlayer(tx, player, state, now); err != nil {
			return err
		}
		result.Save, result.Revision = state, player.Revision
		result.Room, err = gachaRoomView(tx, userID, &room)
		if err != nil {
			return err
		}
		result.Queue, err = gachaArenaQueueView(tx, userID, &seat)
		if err != nil {
			return err
		}
		result.Arena = result.Room.Match
		return recordGachaAction(tx, userID, intent.RequestID, hash, result, now)
	})
	return result, err
}

func GetGachaArenaRooms(userID int, now time.Time) (*GachaArenaLobby, error) {
	if userID <= 0 {
		return nil, ErrGachaIntent
	}
	lobby := &GachaArenaLobby{}
	err := gachaArenaTransaction(func(tx *gorm.DB, deltas map[int]int) error {
		*lobby = GachaArenaLobby{Rooms: []GachaArenaRoom{}}
		var user User
		if err := tx.Select("id", "quota", "status").First(&user, userID).Error; err != nil {
			return err
		}
		if user.Status != common.UserStatusEnabled {
			return ErrGachaIntent
		}
		var seat GachaArenaSeat
		if err := tx.First(&seat, "user_id = ?", userID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		lobby.Queued = seat.RoomID == "" && seat.Status == "waiting"
		if seat.RoomID != "" {
			var room GachaArenaRoom
			if err := tx.First(&room, "id = ?", seat.RoomID).Error; err != nil {
				return err
			}
			if room.Status == "open" && (room.ExpireAt <= now.UnixMilli() || !GachaArenaEnabled()) {
				if err := closeGachaRoom(tx, room.ID, "expired", deltas); err != nil {
					return err
				}
				room.Status = "expired"
			}
			var err error
			lobby.Room, err = gachaRoomView(tx, userID, &room)
			if err != nil {
				return err
			}
		}
		if err := tx.Model(&User{}).Where("id = ?", userID).Select("quota").Scan(&lobby.Quota).Error; err != nil {
			return err
		}
		if !GachaArenaEnabled() {
			return nil
		}
		return tx.Model(&GachaArenaRoom{}).Select("gacha_arena_rooms.*").Joins("JOIN users ON users.id = gacha_arena_rooms.host_id").
			Where("gacha_arena_rooms.status = ? AND gacha_arena_rooms.guest_id = 0 AND gacha_arena_rooms.expire_at > ? AND users.status = ? AND users.deleted_at IS NULL", "open", now.UnixMilli(), common.UserStatusEnabled).
			Order("gacha_arena_rooms.created_at DESC").Limit(50).Find(&lobby.Rooms).Error
	})
	return lobby, err
}
