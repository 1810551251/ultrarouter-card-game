package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const GachaArenaQueueTTL = 2 * time.Minute
const GachaArenaQuotaPerE = 500_000

var (
	ErrGachaArenaWaiting = errors.New("already queued for a match")
	ErrGachaArenaFunds   = errors.New("insufficient available wallet balance")
)

// A shared SQL gate serializes pairing, cancellation and expiry across nodes.
// The UPDATE also acquires SQLite's write lock before any snapshot is read.
type GachaArenaGate struct {
	ID       int   `gorm:"primaryKey;autoIncrement:false"`
	Revision int64 `gorm:"not null"`
}

// A waiting seat owns exactly StakeE of escrow. Only a committed transition to
// matched/cancelled/expired may release it. The ticket prevents stale cancels.
type GachaArenaSeat struct {
	UserID    int    `gorm:"primaryKey;autoIncrement:false"`
	ID        string `gorm:"type:varchar(36);not null;uniqueIndex"`
	Status    string `gorm:"type:varchar(16);not null;index:idx_gacha_queue,priority:1"`
	StakeE    int    `gorm:"not null;index:idx_gacha_queue,priority:2"`
	Power     int    `gorm:"not null;index:idx_gacha_queue,priority:3"`
	CreatedAt int64
	ExpireAt  int64         `gorm:"index"`
	Snapshot  GachaDocument `gorm:"not null"`
	MatchID   string        `gorm:"type:varchar(36)"`
	RoomID    string        `gorm:"type:varchar(36);not null;default:'';index"`
}

type GachaArenaDuel struct {
	ID        string        `gorm:"type:varchar(36);primaryKey"`
	LeftID    int           `gorm:"not null;index"`
	RightID   int           `gorm:"not null;index"`
	LeftName  string        `gorm:"type:text;not null"`
	RightName string        `gorm:"type:text;not null"`
	StakeE    int           `gorm:"not null"`
	WinnerID  int           `gorm:"not null"`
	PlayedAt  int64         `gorm:"index"`
	Report    GachaDocument `gorm:"not null"`
}

// Permanent wallet journal survives replacement of a user's latest queue seat.
// One reserve and one release per ticket also fence duplicate economic effects.
type GachaArenaEntry struct {
	QueueID    string `gorm:"type:varchar(36);primaryKey"`
	Phase      string `gorm:"type:varchar(16);primaryKey"`
	UserID     int    `gorm:"not null;index"`
	DeltaQuota int    `gorm:"not null"`
	Reason     string `gorm:"type:varchar(16);not null"`
	CreatedAt  int64
}

type GachaArenaQueueView struct {
	ID       string            `json:"id"`
	Status   string            `json:"status"`
	StakeE   int               `json:"stake_e"`
	ExpireAt int64             `json:"expire_at"`
	Power    int               `json:"power"`
	MinPower int               `json:"min_power"`
	MaxPower int               `json:"max_power"`
	Quota    int               `json:"quota"`
	Match    *GachaArenaResult `json:"match,omitempty"`
	RoomID   string            `json:"room_id,omitempty"`
}

func GachaArenaEnabled() bool {
	// Batch relay deductions may still be ahead of SQL. Do not reserve game
	// funds against that stale snapshot or change the site's billing settings.
	return common.GetEnvOrDefaultBool("GACHA_MATCHMAKING_ENABLED", true) && !common.BatchUpdateEnabled
}

func InitializeGachaArena() error {
	return DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&GachaArenaGate{ID: 1}).Error
}

// max(a,b) <= min(a,b)*1.2, using integers and inclusive bounds.
// This is symmetric: the weaker side defines the 20% ceiling.
func gachaArenaPowerRange(power int) (int, int) {
	return int((int64(power)*5 + 5) / 6), int(int64(power) * 6 / 5)
}

func lockGachaArena(tx *gorm.DB) error {
	r := tx.Model(&GachaArenaGate{}).Where("id = ?", 1).Update("revision", gorm.Expr("revision + 1"))
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected != 1 {
		return ErrGachaRunState
	}
	return nil
}

// Every wallet change is in the same transaction as its escrow transition.
// Cache deltas are applied only for a committed attempt, never on a retry.
func gachaArenaTransaction(run func(*gorm.DB, map[int]int) error) error {
	var deltas map[int]int
	err := gachaTransaction(func(tx *gorm.DB) error {
		deltas = map[int]int{}
		if err := lockGachaArena(tx); err != nil {
			return err
		}
		return run(tx, deltas)
	})
	if err == nil {
		for id, delta := range deltas {
			if delta != 0 {
				if cacheErr := cacheIncrUserQuota(id, int64(delta)); cacheErr != nil {
					common.SysError(fmt.Sprintf("gacha escrow cache sync failed for user %d: %v", id, cacheErr))
				}
			}
		}
	}
	return err
}

func refundGachaArenaSeat(tx *gorm.DB, seat *GachaArenaSeat, status string, deltas map[int]int) error {
	if seat.Status != "waiting" {
		return nil
	}
	if seat.StakeE < 1 || seat.StakeE > 5 {
		return ErrGachaRunState
	}
	amount := seat.StakeE * GachaArenaQuotaPerE
	// Returning the owner's escrow must also work after an account is disabled.
	r := tx.Unscoped().Model(&User{}).Where("id = ? AND quota <= ? AND quota >= ?", seat.UserID, common.MaxWalletQuota-amount, -common.MaxWalletQuota).
		Update("quota", gorm.Expr("quota + ?", amount))
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected != 1 {
		return ErrGachaCapacity
	}
	r = tx.Model(&GachaArenaSeat{}).Where("user_id = ? AND id = ? AND status = ?", seat.UserID, seat.ID, "waiting").Updates(map[string]any{"status": status, "snapshot": ""})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected != 1 {
		return ErrGachaConcurrent
	}
	seat.Status, seat.Snapshot = status, ""
	deltas[seat.UserID] += amount
	return tx.Create(&GachaArenaEntry{QueueID: seat.ID, Phase: "release", UserID: seat.UserID, DeltaQuota: amount, Reason: status, CreatedAt: time.Now().UnixMilli()}).Error
}

func applyGachaQueueAction(userID int, intent GachaActionIntent, now time.Time) (*GachaActionResult, error) {
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
		err = tx.First(&seat, "user_id = ?", userID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if intent.Action == "arena_cancel" {
			if seat.RoomID != "" {
				return ErrGachaIntent
			}
			if seat.ID != strings.ToLower(*intent.QueueID) {
				return ErrGachaIntent
			}
			if err := refundGachaArenaSeat(tx, &seat, "cancelled", deltas); err != nil {
				return err
			}
		} else {
			if !GachaArenaEnabled() {
				return ErrGachaArenaPending
			}
			if seat.Status == "waiting" || seat.Status == "room" {
				return ErrGachaArenaWaiting
			}
			if state.Daily.ArenaPlayed >= GachaArenaDailyLimit {
				return ErrGachaArenaLimit
			}
			power, err := GachaSquadPower(state)
			if err != nil || power <= 0 {
				return ErrGachaTeam
			}
			amount := *intent.StakeE * GachaArenaQuotaPerE
			r := tx.Model(&User{}).Where("id = ? AND status = ? AND quota >= ? AND quota <= ?", userID, common.UserStatusEnabled, amount, common.MaxWalletQuota-amount).
				Update("quota", gorm.Expr("quota - ?", amount))
			if r.Error != nil {
				return r.Error
			}
			if r.RowsAffected != 1 {
				return ErrGachaArenaFunds
			}
			deltas[userID] -= amount
			snapshot, err := freezeGachaRoster(state)
			if err != nil {
				return err
			}
			seat = GachaArenaSeat{UserID: userID, ID: uuid.NewString(), Status: "waiting", StakeE: *intent.StakeE, Power: power, CreatedAt: now.UnixMilli(), ExpireAt: now.Add(GachaArenaQueueTTL).UnixMilli(), Snapshot: GachaDocument(snapshot)}
			if err := tx.Save(&seat).Error; err != nil {
				return err
			}
			if err := tx.Create(&GachaArenaEntry{QueueID: seat.ID, Phase: "reserve", UserID: userID, DeltaQuota: -amount, Reason: "waiting", CreatedAt: now.UnixMilli()}).Error; err != nil {
				return err
			}
			var other GachaArenaSeat
			minimum, maximum := gachaArenaPowerRange(power)
			err = tx.Model(&GachaArenaSeat{}).Select("gacha_arena_seats.*").Joins("JOIN users ON users.id = gacha_arena_seats.user_id").
				Where("gacha_arena_seats.room_id = ?", "").
				Where("gacha_arena_seats.status = ? AND gacha_arena_seats.stake_e = ? AND gacha_arena_seats.user_id <> ? AND gacha_arena_seats.expire_at > ? AND users.status = ? AND users.deleted_at IS NULL", "waiting", seat.StakeE, userID, now.UnixMilli(), common.UserStatusEnabled).
				Where("gacha_arena_seats.power >= ? AND gacha_arena_seats.power <= ?", minimum, maximum).
				Clauses(clause.OrderBy{Expression: clause.Expr{SQL: "ABS(gacha_arena_seats.power - ?) ASC, gacha_arena_seats.created_at ASC, gacha_arena_seats.user_id ASC", Vars: []any{power}}}).
				Take(&other).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err == nil {
				if err := settleGachaArenaPair(tx, state, &seat, &other, now, deltas); err != nil {
					return err
				}
			}
		}
		if err := storeGachaPlayer(tx, player, state, now); err != nil {
			return err
		}
		result.Save, result.Revision = state, player.Revision
		result.Queue, err = gachaArenaQueueView(tx, userID, &seat)
		if err != nil {
			return err
		}
		result.Arena = result.Queue.Match
		return recordGachaAction(tx, userID, intent.RequestID, hash, result, now)
	})
	return result, err
}

func settleGachaArenaPair(tx *gorm.DB, state *GachaGameState, seat, other *GachaArenaSeat, now time.Time, deltas map[int]int) error {
	if seat.UserID == other.UserID || seat.Status != "waiting" || other.Status != "waiting" || seat.StakeE != other.StakeE || other.ExpireAt <= now.UnixMilli() || seat.ExpireAt <= now.UnixMilli() || seat.RoomID != other.RoomID {
		return ErrGachaIntent
	}
	otherPlayer, otherState, err := loadGachaPlayer(tx, other.UserID, now)
	if err != nil {
		return err
	}
	if state.Daily.ArenaPlayed >= GachaArenaDailyLimit || otherState.Daily.ArenaPlayed >= GachaArenaDailyLimit {
		return ErrGachaArenaLimit
	}
	frozen, err := thawGachaRoster(other.Snapshot)
	if err != nil {
		return err
	}
	minimum, maximum := gachaArenaPowerRange(seat.Power)
	frozenPower, err := GachaSquadPower(frozen)
	if err != nil || frozenPower != other.Power || frozenPower < minimum || frozenPower > maximum {
		return ErrGachaOpponent
	}
	ownFrozen, err := thawGachaRoster(seat.Snapshot)
	if err != nil {
		return err
	}
	ownPower, err := GachaSquadPower(ownFrozen)
	if err != nil || ownPower != seat.Power {
		return ErrGachaOpponent
	}
	report, err := SimulateGachaArena(frozen, ownFrozen)
	if err != nil {
		return err
	}
	users := []User{}
	if err := lockForUpdate(tx).Select("id", "username", "display_name", "status", "quota").Where("id IN ?", []int{seat.UserID, other.UserID}).Order("id ASC").Find(&users).Error; err != nil {
		return err
	}
	if len(users) != 2 {
		return ErrGachaOpponent
	}
	names := map[int]string{}
	for _, user := range users {
		if user.Status != common.UserStatusEnabled {
			return ErrGachaOpponent
		}
		names[user.Id] = user.DisplayName
		if names[user.Id] == "" {
			names[user.Id] = user.Username
		}
	}
	duel := GachaArenaDuel{ID: uuid.NewString(), LeftID: other.UserID, RightID: seat.UserID, LeftName: names[other.UserID], RightName: names[seat.UserID], StakeE: seat.StakeE, PlayedAt: now.UnixMilli()}
	if !report.TimedOut {
		duel.WinnerID = seat.UserID
		if report.Win {
			duel.WinnerID = other.UserID
		}
	}
	for _, pair := range []struct {
		seat  *GachaArenaSeat
		state *GachaGameState
	}{{other, otherState}, {seat, state}} {
		amount := 0
		if duel.WinnerID == 0 {
			amount = duel.StakeE * GachaArenaQuotaPerE
		} else if duel.WinnerID == pair.seat.UserID {
			amount = 2 * duel.StakeE * GachaArenaQuotaPerE
		}
		if amount > 0 {
			r := tx.Model(&User{}).Where("id = ? AND quota <= ? AND quota >= ?", pair.seat.UserID, common.MaxWalletQuota-amount, -common.MaxWalletQuota).Update("quota", gorm.Expr("quota + ?", amount))
			if r.Error != nil {
				return r.Error
			}
			if r.RowsAffected != 1 {
				return ErrGachaCapacity
			}
			deltas[pair.seat.UserID] += amount
		}
		r := tx.Model(&GachaArenaSeat{}).Where("user_id = ? AND id = ? AND status = ?", pair.seat.UserID, pair.seat.ID, "waiting").Updates(map[string]any{"status": "matched", "match_id": duel.ID, "snapshot": ""})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return ErrGachaConcurrent
		}
		pair.seat.Status, pair.seat.MatchID, pair.seat.Snapshot = "matched", duel.ID, ""
		if err := tx.Create(&GachaArenaEntry{QueueID: pair.seat.ID, Phase: "release", UserID: pair.seat.UserID, DeltaQuota: amount, Reason: "matched", CreatedAt: now.UnixMilli()}).Error; err != nil {
			return err
		}
		match := duel.matchFor(pair.seat.UserID)
		pair.state.Arena.Played++
		pair.state.Daily.ArenaPlayed++
		if match.Win {
			pair.state.Arena.Wins++
		}
		pair.state.Arena.History = append([]GachaArenaMatch{match}, pair.state.Arena.History...)
		if len(pair.state.Arena.History) > 10 {
			pair.state.Arena.History = pair.state.Arena.History[:10]
		}
	}
	data, err := common.Marshal(report)
	if err != nil {
		return err
	}
	duel.Report = GachaDocument(data)
	if err := tx.Create(&duel).Error; err != nil {
		return err
	}
	return storeGachaPlayer(tx, otherPlayer, otherState, now)
}

func (d *GachaArenaDuel) matchFor(userID int) GachaArenaMatch {
	m := GachaArenaMatch{ID: d.ID, OpponentID: d.RightID, Name: d.RightName, Win: d.WinnerID == userID, Draw: d.WinnerID == 0, PlayedAt: d.PlayedAt, StakeE: d.StakeE}
	if userID == d.RightID {
		m.OpponentID, m.Name = d.LeftID, d.LeftName
	}
	if m.Win {
		m.DeltaQuota = d.StakeE * GachaArenaQuotaPerE
	} else if !m.Draw {
		m.DeltaQuota = -d.StakeE * GachaArenaQuotaPerE
	}
	return m
}

func gachaArenaResult(tx *gorm.DB, userID int, id string) (*GachaArenaResult, error) {
	var duel GachaArenaDuel
	if err := tx.Where("id = ? AND (left_id = ? OR right_id = ?)", id, userID, userID).First(&duel).Error; err != nil {
		return nil, ErrGachaIntent
	}
	var report GachaBattleReport
	if err := common.Unmarshal([]byte(duel.Report), &report); err != nil {
		return nil, err
	}
	match := duel.matchFor(userID)
	if userID == duel.RightID {
		// Keep each viewer on the left; remap all references to the same report.
		flip := func(uid string) string {
			if strings.HasPrefix(uid, "a") {
				return "e" + uid[1:]
			}
			if strings.HasPrefix(uid, "e") {
				return "a" + uid[1:]
			}
			return uid
		}
		report.Allies, report.Waves[0] = report.Waves[0], report.Allies
		for i := range report.Allies {
			report.Allies[i].UID = flip(report.Allies[i].UID)
			report.Allies[i].Side = "ally"
		}
		for i := range report.Waves[0] {
			report.Waves[0][i].UID = flip(report.Waves[0][i].UID)
			report.Waves[0][i].Side = "enemy"
		}
		for i := range report.Events {
			e := &report.Events[i]
			e.Src = flip(e.Src)
			for j := range e.Targets {
				e.Targets[j].UID = flip(e.Targets[j].UID)
			}
			for j := range e.Units {
				e.Units[j].UID = flip(e.Units[j].UID)
			}
		}
		report.Win, report.Stars = match.Win, 0
		report.TotalN, report.AliveN = len(report.Allies), 0
		if len(report.Events) > 0 {
			for _, u := range report.Events[len(report.Events)-1].Units {
				if strings.HasPrefix(u.UID, "a") && u.Alive {
					report.AliveN++
				}
			}
		}
	}
	return &GachaArenaResult{GachaArenaMatch: match, Report: &report}, nil
}

func gachaArenaQueueView(tx *gorm.DB, userID int, seat *GachaArenaSeat) (*GachaArenaQueueView, error) {
	var user User
	if err := tx.Select("id", "quota", "status").First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.Status != common.UserStatusEnabled {
		return nil, ErrGachaIntent
	}
	view := &GachaArenaQueueView{ID: seat.ID, Status: seat.Status, StakeE: seat.StakeE, Power: seat.Power, ExpireAt: seat.ExpireAt, Quota: user.Quota}
	view.RoomID = seat.RoomID
	if seat.RoomID != "" && (seat.Status == "waiting" || seat.Status == "room") {
		view.Status = "room"
	}
	view.MinPower, view.MaxPower = gachaArenaPowerRange(seat.Power)
	if seat.ID == "" {
		view.Status = "idle"
	}
	if seat.MatchID != "" {
		var err error
		view.Match, err = gachaArenaResult(tx, userID, seat.MatchID)
		if err != nil {
			return nil, err
		}
	}
	return view, nil
}

func GetGachaArenaQueue(userID int, now time.Time) (*GachaArenaQueueView, error) {
	if userID <= 0 {
		return nil, ErrGachaIntent
	}
	var view *GachaArenaQueueView
	err := gachaArenaTransaction(func(tx *gorm.DB, deltas map[int]int) error {
		var seat GachaArenaSeat
		err := tx.First(&seat, "user_id = ?", userID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if seat.Status == "waiting" && (seat.ExpireAt <= now.UnixMilli() || !GachaArenaEnabled()) {
			if seat.RoomID != "" {
				if err := closeGachaRoom(tx, seat.RoomID, "expired", deltas); err != nil {
					return err
				}
				if err := tx.First(&seat, "user_id = ?", userID).Error; err != nil {
					return err
				}
			} else if err := refundGachaArenaSeat(tx, &seat, "expired", deltas); err != nil {
				return err
			}
		}
		view, err = gachaArenaQueueView(tx, userID, &seat)
		return err
	})
	return view, err
}

func GetGachaArenaReplay(userID int, id string) (*GachaArenaResult, error) {
	if userID <= 0 || !validateGachaUUID(id) {
		return nil, ErrGachaIntent
	}
	var user User
	if err := DB.Select("id", "status").First(&user, userID).Error; err != nil || user.Status != common.UserStatusEnabled {
		return nil, ErrGachaIntent
	}
	return gachaArenaResult(DB, userID, strings.ToLower(id))
}

func HasExpiredGachaArenaSeats(now time.Time) bool {
	query := DB.Model(&GachaArenaSeat{}).Where("status IN ?", []string{"waiting", "room"})
	if GachaArenaEnabled() {
		query = query.Where("expire_at <= ?", now.UnixMilli())
	}
	var count int64
	return query.Limit(1).Count(&count).Error == nil && count > 0
}

func ExpireGachaArenaSeats(ctx context.Context, now time.Time) error {
	query := DB.WithContext(ctx).Where("status IN ?", []string{"waiting", "room"})
	if GachaArenaEnabled() {
		query = query.Where("expire_at <= ?", now.UnixMilli())
	}
	var seats []GachaArenaSeat
	if err := query.Order("expire_at ASC").Limit(50).Find(&seats).Error; err != nil {
		return err
	}
	var failure error
	for _, candidate := range seats {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := gachaArenaTransaction(func(tx *gorm.DB, deltas map[int]int) error {
			tx = tx.WithContext(ctx)
			var seat GachaArenaSeat
			if err := tx.Where("user_id = ?", candidate.UserID).First(&seat).Error; err != nil {
				return err
			}
			if seat.ID != candidate.ID || seat.Status != "waiting" && seat.Status != "room" || GachaArenaEnabled() && seat.ExpireAt > now.UnixMilli() {
				return nil
			}
			if seat.RoomID != "" {
				return closeGachaRoom(tx, seat.RoomID, "expired", deltas)
			}
			return refundGachaArenaSeat(tx, &seat, "expired", deltas)
		})
		failure = errors.Join(failure, err)
	}
	return failure
}
