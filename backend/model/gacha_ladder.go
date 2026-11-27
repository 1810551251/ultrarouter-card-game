package model

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const GachaLadderDailyLimit = 5
const GachaLadderBots = 100

var ErrGachaLadder = errors.New("ranked arena state changed or action unavailable")
var ErrGachaLadderLimit = errors.New("ranked arena daily limit or cooldown reached")

// All rank changes, defense changes and daily settlement share this SQL gate.
// Settling before every rank mutation preserves the actual 21:00 ordering even
// if the host worker was offline: no later rank can overwrite an unpaid day.
type GachaLadderGate struct {
	ID             int   `gorm:"primaryKey;autoIncrement:false"`
	Revision       int64 `gorm:"not null"`
	SettledThrough int64 `gorm:"not null"`
}

type GachaLadderEntry struct {
	ID            int               `gorm:"primaryKey;autoIncrement:false" json:"id"` // user ID; negative IDs are NPCs
	Rank          int               `gorm:"column:position;not null;uniqueIndex" json:"rank"`
	BestRank      int               `json:"best_rank"`
	Name          string            `gorm:"type:varchar(128);not null" json:"name"`
	Power         int               `json:"power"`
	Defense       GachaDocument     `gorm:"not null" json:"-"`
	JoinedAt      int64             `json:"joined_at"`
	Day           string            `gorm:"type:varchar(10)" json:"-"`
	Used          int               `json:"-"`
	CooldownUntil int64             `json:"cooldown_until"`
	Squad         []GachaPublicGirl `gorm:"-" json:"squad"`
}

type GachaLadderMaterials struct {
	R   int `json:"mat_r"`
	SR  int `json:"mat_sr"`
	SSR int `json:"mat_ssr"`
}
type GachaLadderRewardTier struct {
	MaxRank   int                  `json:"max_rank"`
	Materials GachaLadderMaterials `json:"materials"`
}

var gachaLadderRewards = []GachaLadderRewardTier{
	{1, GachaLadderMaterials{120, 60, 20}},
	{3, GachaLadderMaterials{100, 50, 15}},
	{10, GachaLadderMaterials{80, 40, 12}},
	{50, GachaLadderMaterials{60, 30, 8}},
	{100, GachaLadderMaterials{40, 20, 5}},
	{0, GachaLadderMaterials{30, 15, 3}},
}

type GachaLadderMail struct {
	UserID    int                  `gorm:"primaryKey;autoIncrement:false" json:"-"`
	Day       string               `gorm:"type:varchar(10);primaryKey" json:"day"`
	Rank      int                  `json:"rank"`
	Materials GachaLadderMaterials `gorm:"embedded;embeddedPrefix:reward_" json:"materials"`
	SettledAt int64                `json:"settled_at"`
	Claimed   bool                 `gorm:"not null;default:false" json:"claimed"`
}

type GachaLadderBattle struct {
	ID             string        `gorm:"type:varchar(36);primaryKey" json:"id"`
	AttackerID     int           `gorm:"index" json:"attacker_id"`
	DefenderID     int           `gorm:"index" json:"defender_id"`
	AttackerName   string        `gorm:"type:varchar(128)" json:"attacker_name"`
	DefenderName   string        `gorm:"type:varchar(128)" json:"defender_name"`
	AttackerBefore int           `json:"attacker_before"`
	AttackerAfter  int           `json:"attacker_after"`
	DefenderBefore int           `json:"defender_before"`
	DefenderAfter  int           `json:"defender_after"`
	AttackWin      bool          `json:"attack_win"`
	PlayedAt       int64         `gorm:"index" json:"played_at"`
	Report         GachaDocument `gorm:"not null" json:"-"`
}

type GachaLadderResult struct {
	GachaLadderBattle
	Ladder    bool               `json:"ladder"`
	Name      string             `json:"name"`
	Win       bool               `json:"win"`
	OwnBefore int                `json:"own_before"`
	OwnAfter  int                `json:"own_after"`
	Report    *GachaBattleReport `json:"report"`
}

type GachaLadderView struct {
	Me             *GachaLadderEntry       `json:"me"`
	Remaining      int                     `json:"remaining"`
	Limit          int                     `json:"limit"`
	Total          int64                   `json:"total"`
	NextSettlement int64                   `json:"next_settlement"`
	NextReset      int64                   `json:"next_reset"`
	ServerNow      int64                   `json:"server_now"`
	Opponents      []GachaLadderEntry      `json:"opponents"`
	Board          []GachaLadderEntry      `json:"board"`
	History        []GachaLadderBattle     `json:"history"`
	Mail           []GachaLadderMail       `json:"mail"`
	Rewards        []GachaLadderRewardTier `json:"rewards"`
}

func gachaLadderBoundary(now time.Time, hour int) time.Time {
	local := now.In(gameZone)
	boundary := time.Date(local.Year(), local.Month(), local.Day(), hour, 0, 0, 0, gameZone)
	if now.Before(boundary) {
		boundary = boundary.AddDate(0, 0, -1)
	}
	return boundary
}
func GachaLadderReward(rank int) GachaLadderMaterials {
	for _, tier := range gachaLadderRewards {
		if rank > 0 && (tier.MaxRank == 0 || rank <= tier.MaxRank) {
			return tier.Materials
		}
	}
	return GachaLadderMaterials{}
}

// Safe to call again after fresh or upgrade migrations. It never resets ranks.
func InitializeGachaLadder(now time.Time) error {
	return gachaTransaction(func(tx *gorm.DB) error {
		gate := GachaLadderGate{ID: 1, SettledThrough: gachaLadderBoundary(now, 21).UnixMilli()}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&gate).Error; err != nil {
			return err
		}
		if err := lockGachaLadder(tx); err != nil {
			return err
		}
		for rank := 1; rank <= GachaLadderBots; rank++ {
			var count int64
			if err := tx.Model(&GachaLadderEntry{}).Where("id = ?", -rank).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			state := newGachaState(now)
			level := 1 + (GachaLadderBots-rank)*59/GachaLadderBots
			for _, own := range state.Girls {
				own.Level, own.Star = level, 1+(level-1)/15
			}
			entry := GachaLadderEntry{ID: -rank, Rank: rank, BestRank: rank, Name: fmt.Sprintf("星霜守卫 %03d", rank), JoinedAt: now.UnixMilli()}
			if err := gachaLadderDefense(&entry, state); err != nil {
				return err
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func lockGachaLadder(tx *gorm.DB) error {
	r := tx.Model(&GachaLadderGate{}).Where("id = ?", 1).Update("revision", gorm.Expr("revision + 1"))
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected != 1 {
		return ErrGachaLadder
	}
	return nil
}

func settleGachaLadder(tx *gorm.DB, now time.Time) error {
	var inactive []int
	enabled := tx.Model(&User{}).Select("id").Where("status = ?", common.UserStatusEnabled)
	if err := tx.Model(&GachaLadderEntry{}).Where("position > 0 AND id > 0 AND id NOT IN (?)", enabled).Order("position ASC").Pluck("id", &inactive).Error; err != nil {
		return err
	}
	var gate GachaLadderGate
	if err := tx.First(&gate, "id = ?", 1).Error; err != nil {
		return err
	}
	due := gachaLadderBoundary(now, 21)
	if gate.SettledThrough >= due.UnixMilli() {
		return removeGachaLadderSeats(tx, inactive)
	}
	// Ranks cannot have changed since SettledThrough without this settlement.
	var players []GachaLadderEntry
	eligible := tx.Select("id", "position", "joined_at").Where("id > 0 AND position > 0")
	if len(inactive) > 0 {
		eligible = eligible.Where("id NOT IN ?", inactive)
	}
	if err := eligible.Find(&players).Error; err != nil {
		return err
	}
	for at := time.UnixMilli(gate.SettledThrough).In(gameZone).AddDate(0, 0, 1); !at.After(due); at = at.AddDate(0, 0, 1) {
		for _, player := range players {
			if player.JoinedAt > at.UnixMilli() {
				continue
			}
			mail := GachaLadderMail{UserID: player.ID, Day: at.Format("2006-01-02"), Rank: player.Rank,
				Materials: GachaLadderReward(player.Rank), SettledAt: at.UnixMilli()}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&mail).Error; err != nil {
				return err
			}
		}
	}
	if err := tx.Model(&gate).Update("settled_through", due.UnixMilli()).Error; err != nil {
		return err
	}
	return removeGachaLadderSeats(tx, inactive)
}

func removeGachaLadderSeats(tx *gorm.DB, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		if err := tx.Model(&GachaLadderEntry{}).Where("id = ?", id).Updates(map[string]any{"position": -id, "defense": GachaDocument("{}"), "power": 0, "name": ""}).Error; err != nil {
			return err
		}
	}
	var active []GachaLadderEntry
	if err := tx.Select("id", "position", "best_rank").Where("position > 0").Order("position ASC").Find(&active).Error; err != nil {
		return err
	}
	for offset, entry := range active {
		rank := offset + 1
		if entry.Rank != rank {
			if err := tx.Model(&entry).Updates(map[string]any{"position": rank, "best_rank": min(rank, entry.BestRank)}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// Call from a host-owned background job (e.g. every minute). Multiple workers
// are safe. Request-time catch-up is a second line of defense, not a client clock.
func SettleGachaLadder(ctx context.Context, now time.Time) error {
	started := time.Now()
	return gachaTransaction(func(tx *gorm.DB) error {
		tx = tx.WithContext(ctx)
		if err := lockGachaLadder(tx); err != nil {
			return err
		}
		return settleGachaLadder(tx, now.Add(time.Since(started)))
	})
}

func gachaLadderDefense(entry *GachaLadderEntry, state *GachaGameState) error {
	power, err := GachaSquadPower(state)
	if err != nil || power <= 0 {
		return ErrGachaTeam
	}
	data, err := common.Marshal(state)
	if err != nil {
		return err
	}
	entry.Defense, entry.Power = GachaDocument(data), power
	return nil
}
func gachaLadderPublic(entry *GachaLadderEntry) error {
	var state GachaGameState
	if err := common.Unmarshal([]byte(entry.Defense), &state); err != nil {
		return err
	}
	entry.Squad = []GachaPublicGirl{}
	for _, id := range state.Squad {
		if girl := state.Girls[id]; girl != nil {
			entry.Squad = append(entry.Squad, GachaPublicGirl{ID: id, Level: girl.Level, Star: girl.Star})
		}
	}
	return nil
}
func gachaLadderIdentity(tx *gorm.DB, userID int) (string, error) {
	if userID <= 0 {
		return "", ErrGachaIntent
	}
	var user User
	if err := tx.Select("id", "status", "username", "display_name").First(&user, userID).Error; err != nil {
		return "", err
	}
	if user.Status != common.UserStatusEnabled {
		return "", ErrGachaIntent
	}
	if user.DisplayName != "" {
		return user.DisplayName, nil
	}
	return user.Username, nil
}

func GetGachaLadder(userID int, now time.Time) (*GachaLadderView, error) {
	view := &GachaLadderView{}
	started := time.Now()
	err := gachaTransaction(func(tx *gorm.DB) error {
		if err := lockGachaLadder(tx); err != nil {
			return err
		}
		// Include time spent waiting for the gate, including transaction retries.
		// A request queued before 21:00 must not mutate yesterday's ordering after
		// acquiring the gate on the other side of the settlement boundary.
		now := now.Add(time.Since(started))
		if _, err := gachaLadderIdentity(tx, userID); err != nil {
			return err
		}
		if err := settleGachaLadder(tx, now); err != nil {
			return err
		}
		*view = GachaLadderView{Limit: GachaLadderDailyLimit, Remaining: GachaLadderDailyLimit,
			ServerNow: now.UnixMilli(), NextSettlement: gachaLadderBoundary(now, 21).AddDate(0, 0, 1).UnixMilli(),
			NextReset: gachaLadderBoundary(now, 5).AddDate(0, 0, 1).UnixMilli(), Rewards: slices.Clone(gachaLadderRewards),
			Opponents: []GachaLadderEntry{}, Board: []GachaLadderEntry{}, History: []GachaLadderBattle{}, Mail: []GachaLadderMail{}}
		if err := tx.Model(&GachaLadderEntry{}).Where("position > 0").Count(&view.Total).Error; err != nil {
			return err
		}
		if err := tx.Where("position > 0").Order("position ASC").Limit(50).Find(&view.Board).Error; err != nil {
			return err
		}
		var me GachaLadderEntry
		err := tx.First(&me, "id = ? AND position > 0", userID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			view.Me = &me
			if me.Day == gachaLadderBoundary(now, 5).Format("2006-01-02") {
				view.Remaining = max(0, GachaLadderDailyLimit-me.Used)
			}
			if err := gachaLadderPublic(&me); err != nil {
				return err
			}
			// Server-chosen rank bands; refresh samples one entry per band.
			low := max(1, me.Rank*40/100)
			high := min(int(view.Total), max(6, me.Rank+2))
			count := high - low
			bands := min(5, count)
			positions := make([]int, 0, bands)
			for band := 0; band < bands; band++ {
				start, end := band*count/bands, (band+1)*count/bands
				pick, err := gachaRandom(end - start)
				if err != nil {
					return err
				}
				position := low + start + pick
				if position >= me.Rank {
					position++
				}
				positions = append(positions, position)
			}
			if len(positions) > 0 {
				if err := tx.Where("position IN ?", positions).Order("position ASC").Find(&view.Opponents).Error; err != nil {
					return err
				}
			}
		}
		for _, entries := range [][]GachaLadderEntry{view.Board, view.Opponents} {
			for i := range entries {
				if err := gachaLadderPublic(&entries[i]); err != nil {
					return err
				}
			}
		}
		if err := tx.Select("id", "attacker_id", "defender_id", "attacker_name", "defender_name", "attacker_before", "attacker_after", "defender_before", "defender_after", "attack_win", "played_at").Where("attacker_id = ? OR defender_id = ?", userID, userID).Order("played_at DESC, id DESC").Limit(20).Find(&view.History).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Order("claimed ASC, day DESC").Limit(60).Find(&view.Mail).Error
	})
	return view, err
}

func validateGachaLadderIntent(intent GachaActionIntent) error {
	if intent.GirlID != nil || intent.Squad != nil || intent.Enabled != nil || intent.StakeE != nil || intent.QueueID != nil || intent.RoomID != nil || intent.Choice != nil || intent.Config != nil {
		return ErrGachaIntent
	}
	switch intent.Action {
	case "ladder_join", "ladder_defense", "ladder_leave":
		if intent.OpponentID != nil || intent.Count != nil || intent.ItemID != nil {
			return ErrGachaIntent
		}
	case "ladder_challenge":
		if intent.OpponentID == nil || *intent.OpponentID == 0 || intent.Count == nil || *intent.Count < 1 || intent.ItemID != nil {
			return ErrGachaIntent
		}
	case "ladder_claim":
		if intent.OpponentID != nil || intent.Count != nil || intent.ItemID == nil {
			return ErrGachaIntent
		}
		if day, err := time.Parse("2006-01-02", *intent.ItemID); err != nil || day.Format("2006-01-02") != *intent.ItemID {
			return ErrGachaIntent
		}
	default:
		return ErrGachaIntent
	}
	return nil
}

func applyGachaLadderAction(userID int, intent GachaActionIntent, now time.Time) (*GachaActionResult, error) {
	intent.RequestID = strings.ToLower(intent.RequestID)
	hash, err := gachaIntentHash(intent)
	if err != nil {
		return nil, err
	}
	result := &GachaActionResult{}
	started := time.Now()
	err = gachaTransaction(func(tx *gorm.DB) error {
		if err := lockGachaLadder(tx); err != nil {
			return err
		}
		now := now.Add(time.Since(started))
		name, err := gachaLadderIdentity(tx, userID)
		if err != nil {
			return err
		}
		if err := settleGachaLadder(tx, now); err != nil {
			return err
		}
		player, state, err := loadGachaPlayer(tx, userID, now)
		if err != nil {
			return err
		}
		if replay, err := gachaReplayResult(tx, userID, intent.RequestID, hash, result); replay || err != nil {
			return err
		}
		*result = GachaActionResult{}
		var me GachaLadderEntry
		err = tx.First(&me, "id = ?", userID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		active := err == nil && me.Rank > 0
		switch intent.Action {
		case "ladder_join":
			if active {
				return ErrGachaLadder
			}
			var last GachaLadderEntry
			if err := tx.Select("position").Order("position DESC").First(&last).Error; err != nil {
				return err
			}
			me.ID, me.Rank, me.Name, me.JoinedAt = userID, last.Rank+1, name, now.UnixMilli()
			if me.BestRank == 0 {
				me.BestRank = me.Rank
			}
			if err := gachaLadderDefense(&me, state); err != nil {
				return err
			}
			if err := tx.Save(&me).Error; err != nil {
				return err
			}
		case "ladder_leave":
			if !active {
				return ErrGachaLadder
			}
			if err := removeGachaLadderSeats(tx, []int{userID}); err != nil {
				return err
			}
		case "ladder_defense":
			if !active {
				return ErrGachaLadder
			}
			me.Name = name
			if err := gachaLadderDefense(&me, state); err != nil {
				return err
			}
			if err := tx.Save(&me).Error; err != nil {
				return err
			}
		case "ladder_challenge":
			if !active {
				return ErrGachaLadder
			}
			result.Ladder, err = challengeGachaLadder(tx, &me, state, *intent.OpponentID, *intent.Count, now)
			if err != nil {
				return err
			}
		case "ladder_claim":
			var mail GachaLadderMail
			if err := tx.First(&mail, "user_id = ? AND day = ?", userID, *intent.ItemID).Error; err != nil || mail.Claimed {
				return ErrGachaLadder
			}
			for key, amount := range map[string]int{"mat_r": mail.Materials.R, "mat_sr": mail.Materials.SR, "mat_ssr": mail.Materials.SSR} {
				if amount < 0 || state.Mats[key] < 0 || state.Mats[key] > gachaMaxResource-amount {
					return ErrGachaCapacity
				}
				state.Mats[key] += amount
			}
			if err := tx.Model(&mail).Update("claimed", true).Error; err != nil {
				return err
			}
		}
		if err := storeGachaPlayer(tx, player, state, now); err != nil {
			return err
		}
		result.Save, result.Revision = state, player.Revision
		return recordGachaAction(tx, userID, intent.RequestID, hash, result, now)
	})
	return result, err
}

func challengeGachaLadder(tx *gorm.DB, me *GachaLadderEntry, attack *GachaGameState, opponentID, expectedRank int, now time.Time) (*GachaLadderResult, error) {
	if opponentID == me.ID {
		return nil, ErrGachaIntent
	}
	day := gachaLadderBoundary(now, 5).Format("2006-01-02")
	if me.Day != day {
		me.Day, me.Used = day, 0
	}
	if me.Used >= GachaLadderDailyLimit || now.UnixMilli() < me.CooldownUntil {
		return nil, ErrGachaLadderLimit
	}
	var opponent GachaLadderEntry
	if err := tx.First(&opponent, "id = ?", opponentID).Error; err != nil {
		return nil, ErrGachaLadder
	}
	// Clients can challenge only the same neighborhood offered by the view.
	if opponent.Rank != expectedRank || opponent.Rank < max(1, me.Rank*40/100) || opponent.Rank > max(6, me.Rank+2) {
		return nil, ErrGachaLadder
	}
	if opponentID > 0 {
		if _, err := gachaLadderIdentity(tx, opponentID); err != nil {
			return nil, ErrGachaLadder
		}
	}
	var defense GachaGameState
	if err := common.Unmarshal([]byte(opponent.Defense), &defense); err != nil {
		return nil, err
	}
	report, err := SimulateGachaArena(attack, &defense)
	if err != nil {
		return nil, err
	}
	battle := GachaLadderBattle{ID: uuid.NewString(), AttackerID: me.ID, DefenderID: opponent.ID,
		AttackerName: me.Name, DefenderName: opponent.Name, AttackerBefore: me.Rank, DefenderBefore: opponent.Rank,
		AttackWin: report.Win && !report.TimedOut, PlayedAt: now.UnixMilli()}
	if battle.AttackWin && opponent.Rank < me.Rank {
		// Temporary rank zero avoids unique-index collisions in every dialect.
		if err := tx.Model(&GachaLadderEntry{}).Where("id = ?", me.ID).Update("position", 0).Error; err != nil {
			return nil, err
		}
		me.Rank, opponent.Rank = opponent.Rank, me.Rank
		me.BestRank = min(me.BestRank, me.Rank)
		if err := tx.Save(&opponent).Error; err != nil {
			return nil, err
		}
	}
	me.Used++
	me.CooldownUntil = now.Add(10 * time.Second).UnixMilli()
	if err := tx.Save(me).Error; err != nil {
		return nil, err
	}
	battle.AttackerAfter, battle.DefenderAfter = me.Rank, opponent.Rank
	data, err := common.Marshal(report)
	if err != nil {
		return nil, err
	}
	battle.Report = GachaDocument(data)
	if err := tx.Create(&battle).Error; err != nil {
		return nil, err
	}
	return gachaLadderResult(battle, me.ID)
}

func gachaLadderResult(battle GachaLadderBattle, userID int) (*GachaLadderResult, error) {
	var report GachaBattleReport
	if err := common.Unmarshal([]byte(battle.Report), &report); err != nil {
		return nil, err
	}
	result := &GachaLadderResult{GachaLadderBattle: battle, Ladder: true, Name: battle.DefenderName, Win: battle.AttackWin,
		OwnBefore: battle.AttackerBefore, OwnAfter: battle.AttackerAfter, Report: &report}
	if userID == battle.DefenderID {
		result.Name, result.Win, result.OwnBefore, result.OwnAfter = battle.AttackerName, !battle.AttackWin, battle.DefenderBefore, battle.DefenderAfter
	}
	return result, nil
}
func GetGachaLadderReplay(userID int, id string) (*GachaLadderResult, error) {
	if !validateGachaUUID(id) {
		return nil, ErrGachaIntent
	}
	if _, err := gachaLadderIdentity(DB, userID); err != nil {
		return nil, err
	}
	var battle GachaLadderBattle
	if err := DB.Where("id = ? AND (attacker_id = ? OR defender_id = ?)", strings.ToLower(id), userID, userID).First(&battle).Error; err != nil {
		return nil, ErrGachaIntent
	}
	return gachaLadderResult(battle, userID)
}
