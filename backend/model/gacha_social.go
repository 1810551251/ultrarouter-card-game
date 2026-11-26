package model

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const GachaArenaDailyLimit = 5
const GachaShopDailyLimit = 20
const gachaPowerVersion = 8

var (
	ErrGachaShopLimit           = errors.New("daily shop limit reached")
	ErrGachaCapacity            = errors.New("resource capacity would be exceeded")
	ErrGachaArenaLimit          = errors.New("daily arena limit reached")
	ErrGachaOpponent            = errors.New("arena opponent is unavailable")
	ErrGachaArenaPending        = errors.New("arena is not open")
	ErrGachaWalletFunds         = errors.New("insufficient wallet balance for token exchange")
	ErrGachaExchangeUnavailable = errors.New("token exchange requires immediate wallet billing")
)

type GachaShopItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Amount   int    `json:"amount"`
	Price    int    `json:"price"`
	Limit    int    `json:"limit"`
}

var gachaShopItems = []GachaShopItem{
	{ID: "gold", Name: "金币补给", Resource: "gold", Amount: 150, Price: 30, Limit: GachaShopDailyLimit},
	{ID: "stamina", Name: "体力补给", Resource: "stamina", Amount: 30, Price: 20, Limit: GachaShopDailyLimit},
	{ID: "mat_r", Name: "微光结晶", Resource: "mat_r", Amount: 20, Price: 15, Limit: GachaShopDailyLimit},
	{ID: "mat_sr", Name: "元素结晶", Resource: "mat_sr", Amount: 10, Price: 25, Limit: GachaShopDailyLimit},
	{ID: "mat_ssr", Name: "星屑", Resource: "mat_ssr", Amount: 5, Price: 40, Limit: GachaShopDailyLimit},
	{ID: "multi", Name: "副本双倍卡", Resource: "multi", Amount: 1, Price: 80, Limit: GachaShopDailyLimit},
}

func gachaShopItem(id string) *GachaShopItem {
	for i := range gachaShopItems {
		if gachaShopItems[i].ID == id {
			return &gachaShopItems[i]
		}
	}
	return nil
}

func buyGachaShopItem(state *GachaGameState, id string) error {
	item := gachaShopItem(id)
	if item == nil {
		return ErrGachaIntent
	}
	remaining := GachaShopDailyLimit
	for _, bought := range state.Daily.ShopBought {
		if bought < 0 {
			return ErrGachaRunState
		}
		if bought >= remaining {
			return ErrGachaShopLimit
		}
		remaining -= bought
	}
	if state.Token < item.Price {
		return ErrGachaFunds
	}
	current, capacity := 0, gachaMaxResource
	switch item.Resource {
	case "gold":
		current = state.Gold
	case "stamina":
		current, capacity = state.Stamina, state.StaminaMax
	case "multi":
		current = state.Multi
	default:
		current = state.Mats[item.Resource]
	}
	if current > capacity-item.Amount {
		return ErrGachaCapacity
	}
	state.Token -= item.Price
	switch item.Resource {
	case "gold":
		state.Gold += item.Amount
	case "stamina":
		state.Stamina += item.Amount
	case "multi":
		state.Multi += item.Amount
	default:
		state.Mats[item.Resource] += item.Amount
	}
	if state.Daily.ShopBought == nil {
		state.Daily.ShopBought = map[string]int{}
	}
	state.Daily.ShopBought[id]++
	return nil
}

type GachaTokenExchangeRules struct {
	Enabled       bool `json:"enabled"`
	QuotaPerToken int  `json:"quota_per_token"`
	MaxCount      int  `json:"max_count"`
}

type GachaTokenExchangeReceipt struct {
	Count       int `json:"count"`
	CostQuota   int `json:"cost_quota"`
	WalletQuota int `json:"wallet_quota"`
}

func exchangeGachaTokens(tx *gorm.DB, userID int, state *GachaGameState, count int) (*GachaTokenExchangeReceipt, error) {
	if common.BatchUpdateEnabled {
		return nil, ErrGachaExchangeUnavailable
	}
	if count < 1 || count > gachaMaxResource {
		return nil, ErrGachaIntent
	}
	if state.Token > gachaMaxResource-count {
		return nil, ErrGachaCapacity
	}
	cost, err := common.WalletQuotaFromDecimalStrict(decimal.NewFromInt(int64(count)).Mul(decimal.NewFromInt(GachaArenaQuotaPerE)))
	if err != nil {
		return nil, ErrGachaIntent
	}
	// The conditional SQL debit shares the wallet row with relay reservations and
	// PvP escrow. The caller commits the debit, tokens and replay receipt together.
	r := tx.Model(&User{}).Where("id = ? AND status = ? AND quota >= ? AND quota <= ?", userID, common.UserStatusEnabled, cost, common.MaxWalletQuota).
		Update("quota", gorm.Expr("quota - ?", cost))
	if r.Error != nil {
		return nil, r.Error
	}
	if r.RowsAffected != 1 {
		return nil, ErrGachaWalletFunds
	}
	var wallet User
	if err := lockForUpdate(tx).Select("quota").First(&wallet, userID).Error; err != nil {
		return nil, err
	}
	state.Token += count
	return &GachaTokenExchangeReceipt{Count: count, CostQuota: cost, WalletQuota: wallet.Quota}, nil
}

type GachaArenaMatch struct {
	ID         string `json:"id"`
	OpponentID int    `json:"opponent_id"`
	Name       string `json:"name"`
	Win        bool   `json:"win"`
	Draw       bool   `json:"draw"`
	Points     int    `json:"points"`
	PlayedAt   int64  `json:"played_at"`
	StakeE     int    `json:"stake_e"`
	DeltaQuota int    `json:"delta_quota"`
}
type GachaArenaState struct {
	Points  int               `json:"points"`
	Wins    int               `json:"wins"`
	Played  int               `json:"played"`
	History []GachaArenaMatch `json:"history"`
}
type GachaArenaResult struct {
	GachaArenaMatch
	Report *GachaBattleReport `json:"report"`
}
type GachaPublicGirl struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
	Star  int    `json:"star"`
}
type GachaCompetitor struct {
	UserID      int               `json:"user_id"`
	Name        string            `json:"name"`
	Power       int               `json:"power"`
	ArenaPoints int               `json:"arena_points"`
	Squad       []GachaPublicGirl `json:"squad" gorm:"-"`
	State       string            `json:"-"`
}
type GachaSocialView struct {
	ArenaEnabled   bool                    `json:"arena_enabled"`
	Shop           []GachaShopItem         `json:"shop"`
	PowerBoard     []GachaCompetitor       `json:"power_board"`
	ArenaBoard     []GachaCompetitor       `json:"arena_board"`
	Opponents      []GachaCompetitor       `json:"opponents"`
	PowerRank      int64                   `json:"power_rank"`
	ArenaRank      int64                   `json:"arena_rank"`
	ArenaLimit     int                     `json:"arena_limit"`
	ShopDailyLimit int                     `json:"shop_daily_limit"`
	TokenExchange  GachaTokenExchangeRules `json:"token_exchange"`
}

// Power includes combat stats and a bounded estimate of each awakening route's utility.
func GachaSquadPower(state *GachaGameState) (int, error) {
	if state == nil {
		return 0, ErrGachaTeam
	}
	if len(state.Squad) == 0 {
		return 0, nil
	}
	units, err := gachaSquadUnits(state, "ally", "a")
	if err != nil {
		return 0, err
	}
	power := 0
	for _, u := range units {
		critExpectation := 1 + float64(u.CR)*float64(u.CD)/10_000
		base := float64(u.Atk)*2.2*critExpectation + float64(u.MaxHP)*0.35 + float64(u.Def)*1.6 + float64(u.Spd)*3 + float64(u.EHR)*2 + float64(u.Res)*2
		power += int(math.Round(base * (1 + float64(u.Awakening)*gachaAwakeningPowerWeights[u.Branch])))
	}
	return power, nil
}

// Backfill only derived indexes. A revision condition protects concurrent saves;
// gameplay JSON and resources are never rewritten by this migration.
func InitializeGachaPower() error {
	var cursor int
	for {
		var players []GachaPlayer
		if err := DB.Where("power_version < ? AND user_id > ?", gachaPowerVersion, cursor).Order("user_id ASC").Limit(100).Find(&players).Error; err != nil {
			return err
		}
		if len(players) == 0 {
			return nil
		}
		for _, player := range players {
			var state GachaGameState
			power := 0
			if common.Unmarshal([]byte(player.State), &state) == nil && migrateGachaEffort(&state) == nil {
				power, _ = GachaSquadPower(&state)
			}
			if err := DB.Model(&GachaPlayer{}).Where("user_id = ? AND revision = ? AND power_version < ?", player.UserID, player.Revision, gachaPowerVersion).
				UpdateColumns(map[string]any{"power": power, "power_version": gachaPowerVersion, "arena_points": state.Arena.Points}).Error; err != nil {
				return err
			}
			cursor = player.UserID
		}
	}
}

func gachaCompetitors(tx *gorm.DB) *gorm.DB {
	return tx.Model(&GachaPlayer{}).Joins("JOIN users ON users.id = gacha_players.user_id").
		Where("users.status = ? AND users.deleted_at IS NULL AND gacha_players.power > ?", common.UserStatusEnabled, 0)
}

func GetGachaSocial(userID int) (*GachaSocialView, error) {
	var user User
	if err := DB.Select("id", "status").First(&user, userID).Error; err != nil {
		return nil, err
	}
	if userID <= 0 || user.Status != common.UserStatusEnabled {
		return nil, ErrGachaIntent
	}
	var player GachaPlayer
	if err := DB.First(&player, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	view := &GachaSocialView{Shop: slices.Clone(gachaShopItems), PowerBoard: []GachaCompetitor{}, ArenaBoard: []GachaCompetitor{}, Opponents: []GachaCompetitor{}, ArenaLimit: GachaArenaDailyLimit}
	view.ArenaEnabled = GachaArenaEnabled()
	view.ShopDailyLimit = GachaShopDailyLimit
	view.TokenExchange = GachaTokenExchangeRules{Enabled: !common.BatchUpdateEnabled, QuotaPerToken: GachaArenaQuotaPerE, MaxCount: gachaMaxResource}
	for i, target := range []*[]GachaCompetitor{&view.PowerBoard} {
		query := gachaCompetitors(DB).Select("gacha_players.user_id, gacha_players.power, gacha_players.arena_points, gacha_players.state, COALESCE(NULLIF(users.display_name, ''), users.username) AS name")
		switch i {
		case 0:
			query = query.Order("gacha_players.power DESC, gacha_players.user_id ASC").Limit(50)
		}
		if err := query.Scan(target).Error; err != nil {
			return nil, err
		}
		for j := range *target {
			entry := &(*target)[j]
			var state GachaGameState
			if err := common.Unmarshal([]byte(entry.State), &state); err != nil {
				return nil, err
			}
			entry.Squad = []GachaPublicGirl{}
			for _, id := range state.Squad {
				if own := state.Girls[id]; own != nil {
					entry.Squad = append(entry.Squad, GachaPublicGirl{ID: id, Level: own.Level, Star: own.Star})
				}
			}
			entry.State = ""
		}
	}
	if player.Power > 0 {
		if err := gachaCompetitors(DB).Where("gacha_players.power > ? OR (gacha_players.power = ? AND gacha_players.user_id < ?)", player.Power, player.Power, userID).Count(&view.PowerRank).Error; err != nil {
			return nil, err
		}
		view.PowerRank++
		if player.ArenaPoints > 0 {
			if err := gachaCompetitors(DB).Where("gacha_players.arena_points > ? OR (gacha_players.arena_points = ? AND (gacha_players.power > ? OR (gacha_players.power = ? AND gacha_players.user_id < ?)))", player.ArenaPoints, player.ArenaPoints, player.Power, player.Power, userID).Count(&view.ArenaRank).Error; err != nil {
				return nil, err
			}
			view.ArenaRank++
		}
	}
	return view, nil
}

func SimulateGachaArena(attack, defense *GachaGameState) (*GachaBattleReport, error) {
	var seed [16]byte
	if _, err := cryptorand.Read(seed[:]); err != nil {
		return nil, err
	}
	return simulateGachaArena(attack, defense, rand.New(rand.NewPCG(binary.LittleEndian.Uint64(seed[:8]), binary.LittleEndian.Uint64(seed[8:]))))
}

func simulateGachaArena(attack, defense *GachaGameState, rng *rand.Rand) (*GachaBattleReport, error) {
	allies, err := gachaSquadUnits(attack, "ally", "a")
	if err != nil {
		return nil, ErrGachaTeam
	}
	enemies, err := gachaSquadUnits(defense, "enemy", "e")
	if err != nil {
		return nil, ErrGachaOpponent
	}
	if rng == nil {
		return nil, ErrGachaIntent
	}
	report := &GachaBattleReport{Allies: []GachaBattleUnit{}, Waves: [][]GachaBattleUnit{{}}, Events: []GachaBattleEvent{}}
	for _, u := range allies {
		report.Allies = append(report.Allies, u.GachaBattleUnit)
	}
	for _, u := range enemies {
		report.Waves[0] = append(report.Waves[0], u.GachaBattleUnit)
	}
	b := &gachaCombat{allies: allies, waves: [][]*gachaCombatUnit{enemies}, rng: rng, formula: gachaCatalog.Formula, report: report, pvp: true}
	return b.fight(60), nil
}
