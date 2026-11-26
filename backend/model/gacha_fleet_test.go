package model_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGachaFleetIntentShapes(t *testing.T) {
	charts, unknown := "charts", "balance"
	tier, invalid, yes := 1, 4, true
	for _, intent := range []model.GachaActionIntent{{Action: "guild_listing", Enabled: &yes}, {Action: "guild_supply", ItemID: &charts}, {Action: "guild_supply_claim", Count: &tier}} {
		intent.RequestID = uuid.NewString()
		require.NoError(t, model.ValidateGachaAction(intent))
		intent.Choice = &charts
		require.ErrorIs(t, model.ValidateGachaAction(intent), model.ErrGachaIntent)
	}
	for _, intent := range []model.GachaActionIntent{{Action: "guild_listing"}, {Action: "guild_listing", Enabled: &yes, ItemID: &charts}, {Action: "guild_supply", ItemID: &unknown}, {Action: "guild_supply", ItemID: &charts, Count: &tier}, {Action: "guild_supply_claim", Count: &invalid}} {
		intent.RequestID = uuid.NewString()
		require.ErrorIs(t, model.ValidateGachaAction(intent), model.ErrGachaIntent)
	}
}

func testGachaFleetDatabase(t *testing.T, now time.Time) {
	leader, member, other, lamp, observer := gachaUser(t), gachaUser(t), gachaUser(t), gachaUser(t), gachaUser(t)
	newIntent := func(action string) model.GachaActionIntent {
		return model.GachaActionIntent{RequestID: uuid.NewString(), Action: action}
	}
	item := func(action, id string) model.GachaActionIntent {
		intent := newIntent(action)
		intent.ItemID = &id
		return intent
	}
	apply := func(id int, intent model.GachaActionIntent) *model.GachaActionResult {
		t.Helper()
		result, err := model.ApplyGachaAction(id, intent, now)
		require.NoError(t, err)
		return result
	}
	board := func(id int, key string, start int, date time.Time) *model.GachaFleetBoard {
		t.Helper()
		result, err := model.GetGachaFleet(id, key, start, date)
		require.NoError(t, err)
		return result
	}
	findCode := func(view *model.GachaFleetBoard, code string) int {
		return slices.IndexFunc(view.Entries, func(entry model.GachaFleetEntry) bool { return entry.Code == code })
	}
	yes := true
	created := apply(leader.Id, item("guild_create", "冬港工坊"))
	code := created.Guild.Code
	create := item("guild_create", "夏夜灯塔")
	create.Enabled = &yes
	otherCode := apply(other.Id, create).Guild.Code
	assert.Equal(t, -1, findCode(board(leader.Id, "directory", leader.Id-1, now), code), "existing invitation-only fleets remain private")
	list := newIntent("guild_listing")
	list.Enabled = &yes
	apply(leader.Id, list)
	for _, id := range []int{member.Id, lamp.Id, observer.Id} {
		apply(id, item("guild_join", code))
	}
	list.RequestID = uuid.NewString()
	_, err := model.ApplyGachaAction(member.Id, list, now)
	require.ErrorIs(t, err, model.ErrGachaGuild)
	directory := board(member.Id, "directory", leader.Id-1, now)
	index := findCode(directory, code)
	require.NotEqual(t, -1, index)
	assert.Equal(t, 4, directory.Entries[index].Members)
	assert.True(t, directory.Entries[index].Mine)
	assert.NotContains(t, fmt.Sprint(directory.Entries), leader.Username)
	assert.Equal(t, -1, findCode(board(member.Id, "directory", leader.Id, now), code))
	apply(leader.Id, item("guild_supply", "charts"))
	apply(other.Id, item("guild_supply", "timber"))
	row := board(member.Id, "guild", 0, now)
	index = findCode(row, code)
	require.NotEqual(t, -1, index)
	assert.Equal(t, index+1, findCode(row, otherCode))
	assert.EqualValues(t, index+1, row.MyRank)
	assert.Equal(t, 2, row.MyValue)
	supply := item("guild_supply", "timber")
	first := apply(member.Id, supply)
	replayed := apply(member.Id, supply)
	assert.Equal(t, first.Save.Gold, replayed.Save.Gold)
	assert.Equal(t, first.Revision, replayed.Revision)
	assert.Equal(t, 1, first.Guild.Harbor.Lines[1].Done)
	_, err = model.ApplyGachaAction(member.Id, item("guild_supply", "charts"), now)
	require.ErrorIs(t, err, model.ErrGachaGuildLimit)
	_, err = model.ApplyGachaAction(lamp.Id, item("guild_supply", "lamps"), now)
	require.ErrorIs(t, err, model.ErrGachaFunds)
	state := gachaState(t, lamp.Id, now)
	assert.False(t, state.Daily.GuildSupplied)
	state.Mats["mat_r"] = 2
	seedGachaState(t, lamp.Id, state)
	apply(lamp.Id, item("guild_supply", "lamps"))
	claim, tier := newIntent("guild_supply_claim"), 1
	claim.Count = &tier
	_, err = model.ApplyGachaAction(observer.Id, claim, now)
	require.ErrorIs(t, err, model.ErrGachaGuildLimit)
	awarded := apply(member.Id, claim)
	assert.Equal(t, first.Save.Gold+100, awarded.Save.Gold)
	assert.Equal(t, first.Save.Mats["forge_ore"]+2, awarded.Save.Mats["forge_ore"])
	assert.True(t, awarded.Guild.Harbor.Beacons[0].Claimed)
	assert.Equal(t, awarded.Revision, apply(member.Id, claim).Revision)
	claim.RequestID = uuid.NewString()
	_, err = model.ApplyGachaAction(member.Id, claim, now)
	require.ErrorIs(t, err, model.ErrGachaGuildLimit)
	apply(member.Id, newIntent("guild_leave"))
	joined := apply(member.Id, item("guild_join", otherCode))
	assert.True(t, joined.Guild.Harbor.Beacons[0].Claimed)
	assert.Zero(t, joined.Guild.Harbor.MyCargo)
	_, err = model.ApplyGachaAction(member.Id, item("guild_supply", "charts"), now)
	require.ErrorIs(t, err, model.ErrGachaGuildLimit)
	future := now.AddDate(0, 0, 7)
	assert.Zero(t, board(other.Id, "guild", 0, future).MyRank)
	profile, err := model.GetGachaProfile(member.Id, future)
	require.NoError(t, err)
	assert.Zero(t, profile.Guild.Harbor.Score)
	assert.False(t, profile.Guild.Harbor.Beacons[0].Claimed)
	assert.False(t, profile.Guild.Harbor.Supplied)

	displays := []string{fmt.Sprintf("藏家甲%d", leader.Id), fmt.Sprintf("藏家乙%d", other.Id)}
	for index, id := range []int{leader.Id, other.Id} {
		require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", id).Update("display_name", displays[index]).Error)
		state := gachaState(t, id, now)
		for seq := 1; seq <= 27; seq++ {
			state.Girls[fmt.Sprintf("c%02d", seq)] = &model.GachaOwnedGirl{Star: 1, Level: 1}
		}
		state.Girls["c28"] = &model.GachaOwnedGirl{Shards: 1}
		state.Squad = []string{}
		seedGachaState(t, id, state)
		gachaState(t, id, now)
	}
	collection := board(leader.Id, "collection", 0, now)
	index = slices.IndexFunc(collection.Entries, func(entry model.GachaFleetEntry) bool { return entry.Mine })
	require.GreaterOrEqual(t, index, 0)
	require.Less(t, index+1, len(collection.Entries))
	assert.Equal(t, displays[1], collection.Entries[index+1].Name)
	assert.Equal(t, 28, collection.CatalogSize)
	assert.Equal(t, 27, collection.MyValue)
	assert.EqualValues(t, index+1, collection.MyRank)
	assert.Zero(t, board(leader.Id, "power", 0, now).MyRank)
	public, err := common.Marshal(collection)
	require.NoError(t, err)
	for _, field := range []string{"state", "members", "user_id", "email", "quota", "squad", "username", "code"} {
		assert.NotContains(t, string(public), `"`+field+`":`)
	}
	var before, after model.GachaPlayer
	require.NoError(t, model.DB.First(&before, "user_id = ?", other.Id).Error)
	require.NoError(t, model.DB.Model(&before).UpdateColumns(map[string]any{"fleet_indexed": false, "collection_count": 0, "fleet_members": 0}).Error)
	require.NoError(t, model.InitializeGachaPower())
	require.NoError(t, model.InitializeGachaPower())
	require.NoError(t, model.DB.First(&after, "user_id = ?", other.Id).Error)
	assert.Equal(t, before.State, after.State)
	assert.Equal(t, before.Revision, after.Revision)
	assert.Equal(t, before.UpdatedAt, after.UpdatedAt)
	assert.Equal(t, 27, after.CollectionCount)
	assert.Equal(t, 2, after.FleetMembers)
	require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", other.Id).Update("status", common.UserStatusDisabled).Error)
	assert.Equal(t, -1, findCode(board(leader.Id, "directory", leader.Id-1, now), otherCode))
	assert.False(t, slices.ContainsFunc(board(leader.Id, "collection", 0, now).Entries, func(entry model.GachaFleetEntry) bool { return entry.Name == displays[1] }))
	_, err = model.GetGachaFleet(other.Id, "power", 0, now)
	assert.Error(t, err)
}
