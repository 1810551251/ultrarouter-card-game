package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const fleetCargoLimit = 10
const fleetBoardLimit = 50
const fleetDirectoryLimit = 12

type GachaFleetClaims struct {
	Week string `json:"week"`
	Mask int    `json:"mask"`
}

type fleetCargoSpec struct {
	ID, Name, Resource, CostText string
	Cost                         int
}

var fleetCargo = [...]fleetCargoSpec{
	{"charts", "勘绘航标", "stamina", "体力 5", 5},
	{"timber", "修整浮桥", "gold", "金币 45", 45},
	{"lamps", "点亮引航灯", "mat_r", "微光结晶 2", 2},
}

type fleetBeaconSpec struct {
	Name            string
	Goal, Gold, Ore int
}

var fleetBeacons = [...]fleetBeaconSpec{
	{"泊船浅湾", 1, 100, 2}, {"双灯渡口", 3, 180, 4}, {"极星母港", 6, 300, 8},
}

type GachaFleetCargoLine struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Cost     string `json:"cost"`
	Done     int    `json:"done"`
	Eligible bool   `json:"eligible"`
}

type GachaFleetBeacon struct {
	Name     string `json:"name"`
	Tier     int    `json:"tier"`
	Goal     int    `json:"goal"`
	Gold     int    `json:"gold"`
	Ore      int    `json:"ore"`
	Eligible bool   `json:"eligible"`
	Claimed  bool   `json:"claimed"`
}

type GachaFleetHarbor struct {
	Lines     []GachaFleetCargoLine `json:"lines"`
	Beacons   []GachaFleetBeacon    `json:"beacons"`
	LineLimit int                   `json:"line_limit"`
	Score     int                   `json:"score"`
	MyCargo   int                   `json:"my_cargo"`
	Supplied  bool                  `json:"supplied"`
}

func fleetCargoIndex(id string) int {
	for index, rule := range fleetCargo {
		if rule.ID == id {
			return index
		}
	}
	return -1
}

func fleetResource(state *GachaGameState, resource string) int {
	switch resource {
	case "gold":
		return state.Gold
	case "stamina":
		return state.Stamina
	default:
		return state.Mats[resource]
	}
}

func fleetScore(guild *GachaGuildState) int {
	return guild.TotalDamage/100 + (guild.Cargo[0]+guild.Cargo[1]+guild.Cargo[2])*2
}

func fleetHarborView(state *GachaGameState, guild *GachaGuildState, week string) GachaFleetHarbor {
	view := GachaFleetHarbor{LineLimit: fleetCargoLimit, Score: fleetScore(guild), Supplied: state.Daily.GuildSupplied}
	if state.Guild.ContributionWeek == week {
		view.MyCargo = state.Guild.MyCargo
	}
	for index, rule := range fleetCargo {
		view.Lines = append(view.Lines, GachaFleetCargoLine{rule.ID, rule.Name, rule.CostText, guild.Cargo[index],
			!view.Supplied && guild.Cargo[index] < fleetCargoLimit && fleetResource(state, rule.Resource) >= rule.Cost})
	}
	for index, rule := range fleetBeacons {
		claimed := state.FleetClaims != nil && state.FleetClaims.Week == week && state.FleetClaims.Mask&(1<<index) != 0
		ready := min(guild.Cargo[0], guild.Cargo[1], guild.Cargo[2]) >= rule.Goal
		view.Beacons = append(view.Beacons, GachaFleetBeacon{rule.Name, index + 1, rule.Goal, rule.Gold, rule.Ore, ready && !claimed && view.MyCargo > 0, claimed})
	}
	return view
}

func applyFleetCargo(state *GachaGameState, owner *GachaGuildState, intent GachaActionIntent, result *GachaActionResult, week string) error {
	if intent.Action == "guild_supply" {
		index := fleetCargoIndex(*intent.ItemID)
		rule := fleetCargo[index]
		if state.Daily.GuildSupplied || owner.Cargo[index] >= fleetCargoLimit {
			return ErrGachaGuildLimit
		}
		if fleetResource(state, rule.Resource) < rule.Cost {
			return ErrGachaFunds
		}
		switch rule.Resource {
		case "stamina":
			state.Stamina -= rule.Cost
		case "gold":
			state.Gold -= rule.Cost
		default:
			state.Mats[rule.Resource] -= rule.Cost
		}
		state.Daily.GuildSupplied = true
		state.Guild.MyCargo++
		owner.Cargo[index]++
		result.Drops = []string{rule.Name + "已完成，公会航绩 +2"}
		return nil
	}
	index := *intent.Count - 1
	beacon := fleetHarborView(state, owner, week).Beacons[index]
	if !beacon.Eligible {
		return ErrGachaGuildLimit
	}
	if state.Gold > gachaMaxResource-beacon.Gold || state.Mats["forge_ore"] > gachaMaxResource-beacon.Ore {
		return ErrGachaCapacity
	}
	if state.FleetClaims == nil || state.FleetClaims.Week != week {
		state.FleetClaims = &GachaFleetClaims{Week: week}
	}
	state.FleetClaims.Mask |= 1 << index
	state.Gold += beacon.Gold
	state.Mats["forge_ore"] += beacon.Ore
	result.Drops = []string{fmt.Sprintf("%s：金币 +%d · 星锻矿 +%d", beacon.Name, beacon.Gold, beacon.Ore)}
	return nil
}

func fleetPublicName(name string, id int) string {
	if display := strings.TrimSpace(name); display != "" {
		return display
	}
	return fmt.Sprintf("航海员 %04X", id)
}

func gachaFleetIndex(userID int, state *GachaGameState) map[string]any {
	owned := 0
	for _, girl := range gachaCatalog.Girls {
		if gachaCardOwned(state.Girls[girl.ID]) {
			owned++
		}
	}
	columns := map[string]any{"fleet_indexed": true, "collection_count": owned, "fleet_leader": 0,
		"fleet_members": 0, "fleet_week": "", "fleet_score": 0, "fleet_name": ""}
	if guild := state.Guild; guild != nil {
		columns["fleet_leader"] = guild.LeaderID
		if guild.Listed && validGachaGuildOwner(guild, userID, fmt.Sprintf("%08X", userID)) {
			columns["fleet_members"], columns["fleet_name"] = len(guild.Members), guild.Name
			columns["fleet_week"], columns["fleet_score"] = guild.Week, fleetScore(guild)
		}
	}
	return columns
}

func initializeGachaFleetIndexes() error {
	cursor := 0
	for {
		var batch []GachaPlayer
		if err := DB.Select("user_id", "revision", "state").Where("(fleet_indexed = ? OR fleet_indexed IS NULL) AND user_id > ?", false, cursor).Order("user_id ASC").Limit(100).Find(&batch).Error; err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		for _, row := range batch {
			var state GachaGameState
			if err := common.Unmarshal([]byte(row.State), &state); err != nil {
				return err
			}
			if err := DB.Model(&GachaPlayer{}).Where("user_id = ? AND revision = ? AND (fleet_indexed = ? OR fleet_indexed IS NULL)", row.UserID, row.Revision, false).
				UpdateColumns(gachaFleetIndex(row.UserID, &state)).Error; err != nil {
				return err
			}
			cursor = row.UserID
		}
	}
}

type GachaFleetEntry struct {
	Name    string `json:"name"`
	Value   int    `json:"value"`
	Mine    bool   `json:"mine,omitempty"`
	Members int    `json:"members,omitempty"`
	Code    string `json:"code,omitempty"`
}

type GachaFleetBoard struct {
	Board       string            `json:"board"`
	Week        string            `json:"week"`
	Entries     []GachaFleetEntry `json:"entries"`
	MyRank      int64             `json:"my_rank"`
	MyValue     int               `json:"my_value"`
	CatalogSize int               `json:"catalog_size"`
	MemberLimit int               `json:"member_limit"`
	Next        int               `json:"next,omitempty"`
}

func fleetEligible() *gorm.DB {
	return DB.Model(&GachaPlayer{}).Joins("JOIN users ON users.id = gacha_players.user_id").
		Where("users.status = ? AND users.deleted_at IS NULL", common.UserStatusEnabled)
}

func GetGachaFleet(userID int, board string, after int, now time.Time) (*GachaFleetBoard, error) {
	if userID < 1 || after < 0 || board != "directory" && after != 0 {
		return nil, ErrGachaIntent
	}
	week := gachaWeekKey(now)
	column, name := "", "COALESCE(users.display_name, '')"
	criteria := func() *gorm.DB { return fleetEligible() }
	switch board {
	case "directory":
		column, name = "gacha_players.fleet_score", "gacha_players.fleet_name"
		criteria = func() *gorm.DB {
			return fleetEligible().Where("gacha_players.fleet_members BETWEEN 1 AND ?", gachaGuildMemberLimit-1)
		}
	case "guild":
		column, name = "gacha_players.fleet_score", "gacha_players.fleet_name"
		criteria = func() *gorm.DB {
			return fleetEligible().Where("gacha_players.fleet_week = ? AND gacha_players.fleet_members > 0", week)
		}
	case "collection":
		column = "gacha_players.collection_count"
	case "power":
		column = "gacha_players.power"
	default:
		return nil, ErrGachaIntent
	}
	var viewer GachaPlayer
	requester := fleetEligible().Select("gacha_players.user_id", "gacha_players.fleet_leader").Where("gacha_players.user_id = ?", userID).Limit(1).Find(&viewer)
	if requester.Error != nil {
		return nil, requester.Error
	}
	if requester.RowsAffected != 1 {
		return nil, ErrGachaIntent
	}
	view := &GachaFleetBoard{Board: board, Week: week, Entries: []GachaFleetEntry{}, CatalogSize: len(gachaCatalog.Girls), MemberLimit: gachaGuildMemberLimit}
	ownerID := userID
	if board == "directory" || board == "guild" {
		ownerID = viewer.FleetLeader
	}
	var rows []struct {
		UserID, Value, Members int
		Name, Week             string
	}
	projection := "gacha_players.user_id, " + name + " AS name, " + column + " AS value"
	if board == "directory" || board == "guild" {
		projection += ", gacha_players.fleet_members AS members, gacha_players.fleet_week AS week"
	}
	query := criteria().Select(projection)
	if board == "directory" {
		query = query.Where("gacha_players.user_id > ?", after).Order("gacha_players.user_id ASC").Limit(fleetDirectoryLimit + 1)
	} else {
		query = query.Where(column + " > 0").Order(column + " DESC, gacha_players.user_id ASC").Limit(fleetBoardLimit)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	if board == "directory" && len(rows) > fleetDirectoryLimit {
		view.Next = rows[fleetDirectoryLimit-1].UserID
		rows = rows[:fleetDirectoryLimit]
	}
	for index, row := range rows {
		entry := GachaFleetEntry{Name: row.Name, Value: row.Value, Mine: row.UserID == ownerID}
		if board == "directory" || board == "guild" {
			entry.Code, entry.Members = fmt.Sprintf("%08X", row.UserID), row.Members
			if row.Week != week {
				entry.Value = 0
			}
		} else {
			entry.Name = fleetPublicName(row.Name, row.UserID)
		}
		if entry.Mine && board != "directory" {
			view.MyRank, view.MyValue = int64(index+1), entry.Value
		}
		view.Entries = append(view.Entries, entry)
	}
	if board != "directory" && ownerID > 0 && view.MyRank == 0 {
		if err := criteria().Where("gacha_players.user_id = ?", ownerID).Pluck(column, &view.MyValue).Error; err != nil {
			return nil, err
		}
		if view.MyValue > 0 {
			if err := criteria().Where("("+column+" > ? OR ("+column+" = ? AND gacha_players.user_id < ?))", view.MyValue, view.MyValue, ownerID).Count(&view.MyRank).Error; err != nil {
				return nil, err
			}
			view.MyRank++
		}
	}
	return view, nil
}
