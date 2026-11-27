package model_test

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func refinementIntent(action, girl, key string, slot int) model.GachaActionIntent {
	intent := cultivationIntent(action, girl)
	intent.Count, intent.ItemID = &slot, &key
	return intent
}

func TestGachaRefinementCanonicalIntent(t *testing.T) {
	for _, sample := range []struct{ action, base string }{{"gem_refine", "ruby"}, {"chip_refine", "edge_1"}} {
		for _, suffix := range []string{"", "@1", "@12"} {
			require.NoError(t, model.ValidateGachaAction(refinementIntent(sample.action, "c01", sample.base+suffix, 1)))
		}
		for _, suffix := range []string{"@", "@0", "@13", "@01", "@+1", "@-1", "@1@1", "@ 1"} {
			assert.ErrorIs(t, model.ValidateGachaAction(refinementIntent(sample.action, "c01", sample.base+suffix, 1)), model.ErrGachaIntent)
		}
		bad := refinementIntent(sample.action, "c01", sample.base, 1)
		enabled, count, choice, squad := true, 1, "ruby", []string{"c01"}
		for _, alter := range []func(*model.GachaActionIntent){func(intent *model.GachaActionIntent) { intent.Enabled = &enabled }, func(intent *model.GachaActionIntent) { intent.Choice = &choice }, func(intent *model.GachaActionIntent) { intent.StakeE = &count }, func(intent *model.GachaActionIntent) { intent.OpponentID = &count }, func(intent *model.GachaActionIntent) { intent.QueueID = &choice }, func(intent *model.GachaActionIntent) { intent.RoomID = &choice }, func(intent *model.GachaActionIntent) { intent.Squad = &squad }, func(intent *model.GachaActionIntent) { intent.Config = &model.GachaProgressionConfig{} }, func(intent *model.GachaActionIntent) { intent.GirlID = nil }, func(intent *model.GachaActionIntent) { intent.ItemID = nil }, func(intent *model.GachaActionIntent) { intent.Count = nil }} {
			forged := bad
			alter(&forged)
			assert.ErrorIs(t, model.ValidateGachaAction(forged), model.ErrGachaIntent)
		}
	}
	assert.ErrorIs(t, model.ValidateGachaAction(refinementIntent("chip_refine", "c01", "edge_1@1", 2)), model.ErrGachaIntent)
}

func TestGachaEquippedRefinementStatsAndTriggeredEffects(t *testing.T) {
	own := &model.GachaOwnedGirl{Star: 1, Level: 1, Sockets: 2, Gems: []string{"jade", "jade"}}
	state := &model.GachaGameState{Girls: map[string]*model.GachaOwnedGirl{"c01": own}, Squad: []string{"c01"}}
	units := func() (model.GachaBattleUnit, model.GachaBattleUnit) {
		pve, err := model.SimulateGachaCombat(state, model.GachaStage{Waves: 1, EnemyCount: 1, EnemyLv: 1, StarRound: 5})
		require.NoError(t, err)
		arena, err := model.SimulateGachaArena(state, state)
		require.NoError(t, err)
		assert.Equal(t, pve.Allies[0].MaxHP, arena.Allies[0].MaxHP)
		return pve.Allies[0], arena.Waves[0][0]
	}
	base, _ := units()
	own.Gems = []string{"jade@6", "jade@12"}
	first, other := units()
	for _, unit := range []model.GachaBattleUnit{first, other} {
		assert.Equal(t, base.MaxHP+12*(6+12), unit.MaxHP)
		assert.Equal(t, unit.MaxHP*2/100, unit.Shield)
	}
	own.Chips = []string{"tempo_1@12"}
	changed, _ := units()
	assert.Equal(t, first.MaxHP+22+12*7, changed.MaxHP)
	assert.Equal(t, first.Rage+5, changed.Rage)
	state.GemInventory = map[string]int{"jade@12": 100}
	state.ChipInventory = map[string]int{"tempo_1@12": 100}
	same, _ := units()
	assert.Equal(t, changed, same)
}

func runGachaRefinementMatrix(t *testing.T, now time.Time) {
	for _, fixture := range []struct{ kind, base, mat string }{{"gem", "ruby", "gear_thread"}, {"chip", "edge_1", "gear_plate"}} {
		t.Run(fixture.kind+"_refine_replay_transfer_and_exact_item", func(t *testing.T) {
			user := gachaUser(t)
			state := gachaState(t, user.Id, now)
			first, second := "c01", "c08"
			slots := func(state *model.GachaGameState, girl string) []string {
				if fixture.kind == "gem" {
					return state.Girls[girl].Gems
				}
				return state.Girls[girl].Chips
			}
			inventory := func(state *model.GachaGameState) map[string]int {
				if fixture.kind == "gem" {
					return state.GemInventory
				}
				return state.ChipInventory
			}
			state.Girls[first].Sockets, state.Girls[second].Sockets = 1, 1
			inventory(state)[fixture.base] = 2
			state.Mats[fixture.mat], state.Mats["forge_ore"] = 8, 2
			seedGachaState(t, user.Id, state)
			action := func(command, girl, key string) (*model.GachaActionResult, error) {
				intent := refinementIntent(fixture.kind+"_"+command, girl, key, 1)
				if command == "remove" {
					intent.ItemID = nil
				}
				return model.ApplyGachaAction(user.Id, intent, now)
			}
			_, err := action("equip", first, fixture.base)
			require.NoError(t, err)
			intent := refinementIntent(fixture.kind+"_refine", first, fixture.base, 1)
			if fixture.kind == "chip" {
				lock := artifactSlot("chip_lock", first, 1)
				enabled := true
				lock.Enabled = &enabled
				_, err = model.ApplyGachaAction(user.Id, lock, now)
				require.NoError(t, err)
				_, err = model.ApplyGachaAction(user.Id, intent, now)
				require.ErrorIs(t, err, model.ErrGachaCultivationLocked)
				lock = artifactSlot("chip_lock", first, 1)
				enabled = false
				lock.Enabled = &enabled
				_, err = model.ApplyGachaAction(user.Id, lock, now)
				require.NoError(t, err)
			}
			out, err := model.ApplyGachaAction(user.Id, intent, now)
			require.NoError(t, err)
			assert.Equal(t, fixture.base+"@1", slots(out.Save, first)[0])
			assert.Equal(t, 1, inventory(out.Save)[fixture.base])
			assert.Equal(t, 5, out.Save.Mats[fixture.mat])
			assert.Equal(t, 1, out.Save.Mats["forge_ore"])
			replayed, err := model.ApplyGachaAction(user.Id, intent, now)
			require.NoError(t, err)
			assert.Equal(t, out, replayed)
			_, err = action("refine", first, fixture.base)
			require.ErrorIs(t, err, model.ErrGachaCultivationLocked)
			_, err = action("equip", second, fixture.base+"@1")
			require.ErrorIs(t, err, model.ErrGachaFunds)
			out, err = action("equip", first, fixture.base)
			require.NoError(t, err)
			assert.Equal(t, 1, inventory(out.Save)[fixture.base+"@1"])
			assert.Zero(t, inventory(out.Save)[fixture.base])
			assert.Equal(t, fixture.base, slots(out.Save, first)[0])
			out, err = action("equip", second, fixture.base+"@1")
			require.NoError(t, err)
			assert.Zero(t, inventory(out.Save)[fixture.base+"@1"])
			out, err = action("refine", second, fixture.base+"@1")
			require.NoError(t, err)
			assert.Equal(t, fixture.base+"!3.1", slots(out.Save, second)[0])
			assert.Zero(t, out.Save.Mats[fixture.mat])
			assert.Zero(t, out.Save.Mats["forge_ore"])
			_, err = action("refine", second, fixture.base+"!3.1")
			require.ErrorIs(t, err, model.ErrGachaFunds)
			out, err = action("remove", second, "")
			require.NoError(t, err)
			assert.Equal(t, 1, inventory(out.Save)[fixture.base+"!3.1"])
			assert.Equal(t, fixture.base, slots(out.Save, first)[0])
			assert.Empty(t, slots(out.Save, second)[0])
		})
	}
	t.Run("gear_overlimit_requires_both_materials_and_each_star_cost_once", func(t *testing.T) {
		user := gachaUser(t)
		state := gachaState(t, user.Id, now)
		id := "c01"
		own := state.Girls[id]
		own.Star, own.Level, own.Awakening, own.Branch = 12, 60, 3, model.GachaAwakeningBranches(id)[0].ID
		own.Equipment = map[string]model.GachaEquipmentPiece{"head": {Level: 30, Star: 12}}
		state.Mats["gear_rivet"], state.Mats["route_badge"] = 1000, 59
		seedGachaState(t, user.Id, state)
		_, err := model.ApplyGachaAction(user.Id, equipmentIntent("gear_star", id, "head"), now)
		require.ErrorIs(t, err, model.ErrGachaFunds)
		state = gachaState(t, user.Id, now)
		assert.Equal(t, 1000, state.Mats["gear_rivet"])
		state.Mats["route_badge"] = 320
		seedGachaState(t, user.Id, state)
		for index, left := range [][2]int{{770, 260}, {450, 160}, {0, 0}} {
			intent := equipmentIntent("gear_star", id, "head")
			out, err := model.ApplyGachaAction(user.Id, intent, now)
			require.NoError(t, err)
			assert.Equal(t, 13+index, out.Save.Girls[id].Equipment["head"].Star)
			assert.Equal(t, left, [2]int{out.Save.Mats["gear_rivet"], out.Save.Mats["route_badge"]})
			again, err := model.ApplyGachaAction(user.Id, intent, now)
			require.NoError(t, err)
			assert.Equal(t, out.Revision, again.Revision)
		}
		_, err = model.ApplyGachaAction(user.Id, equipmentIntent("gear_star", id, "head"), now)
		require.ErrorIs(t, err, model.ErrGachaCultivationLocked)
	})
}
