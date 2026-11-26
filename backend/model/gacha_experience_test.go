package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGachaLevelCurveAndBulkFeeding(t *testing.T) {
	state := battleTestState([]string{"c01"}, 1, 1)
	state.Captain = &GachaCaptain{Level: gachaCatalog.Growth.MaxLevel}
	own := state.Girls["c01"]
	for level, cost := range map[int]int{0: 100, 1: 100, 59: 5900, 60: 6010, 120: 49235, 149: 118400, 150: 0} {
		assert.Equal(t, cost, gachaLevelCost(level))
	}
	for level := 2; level < gachaCatalog.Growth.MaxLevel; level++ {
		assert.Greater(t, gachaLevelCost(level), gachaLevelCost(level-1))
	}
	own.Level, own.Exp = 149, gachaLevelCost(149)-601
	state.Mats["training_manual"], state.Mats["field_notes"], state.Mats["astral_treatise"] = 1, 2, 4
	intent := trainingTestIntent("card_feed", "")
	intent.Materials = map[string]int{"training_manual": 1, "field_notes": 2, "astral_treatise": 4}
	result := &GachaActionResult{}
	require.NoError(t, ValidateGachaAction(intent))
	require.NoError(t, applyGachaFeed(state, intent, result))
	assert.Equal(t, map[string]int{"training_manual": 1, "field_notes": 1}, result.Feed.Spent)
	assert.Equal(t, 601, result.Feed.Experience)
	assert.Equal(t, 4, state.Mats["astral_treatise"])
	assert.Equal(t, 150, own.Level)
	assert.Zero(t, own.Exp)
	require.ErrorIs(t, applyGachaFeed(state, intent, &GachaActionResult{}), ErrGachaCultivationLocked)
	own.Level = 1
	require.ErrorIs(t, applyGachaFeed(state, intent, &GachaActionResult{}), ErrGachaFunds)
	assert.Equal(t, 1, state.Mats["field_notes"])
	for _, bad := range []map[string]int{nil, {}, {"mat_r": 1}, {"training_manual": 0}, {"field_notes": -1}, {"astral_treatise": gachaMaxResource + 1}} {
		intent.Materials = bad
		require.ErrorIs(t, ValidateGachaAction(intent), ErrGachaIntent)
	}
	intent.Action, intent.Materials = "card_breakthrough", map[string]int{"training_manual": 1}
	require.ErrorIs(t, ValidateGachaAction(intent), ErrGachaIntent)
}

func TestGachaEffortMigrationAndRebornCombat(t *testing.T) {
	state := battleTestState([]string{"c01"}, 60, 5)
	own := state.Girls["c01"]
	own.Training = &GachaCardTraining{Stage: 1, Grade: 1, Effort: map[string]int{"hp": 4, "atk": 3, "def": 2, "spd": 2, "res": 2}}
	require.NoError(t, migrateGachaEffort(state))
	require.NoError(t, migrateGachaEffort(state))
	assert.Equal(t, map[string]int{"hp": 400, "atk": 280, "def": 160, "spd": 160, "res": 160}, own.Training.Effort)
	assert.Equal(t, 60, own.Level)
	require.True(t, validGachaCardTraining(own))
	own.Level, own.Exp, own.PeakLevel = 0, 0, 60
	for _, side := range []string{"ally", "enemy"} {
		units, err := gachaSquadUnits(state, side, "r")
		require.NoError(t, err)
		assert.Zero(t, units[0].Level)
		assert.Positive(t, units[0].MaxHP)
	}
	require.NoError(t, awardGachaExperience(state, state.Squad, 0))
	require.Zero(t, own.Level)
	require.NoError(t, awardGachaExperience(state, state.Squad, 99))
	require.Zero(t, own.Level)
	require.NoError(t, awardGachaExperience(state, state.Squad, 1))
	require.Equal(t, 1, own.Level)
	require.Zero(t, own.Exp)
	assert.False(t, gachaCardOwned(&GachaOwnedGirl{Star: 1, Level: 0}))
	state.EffortVersion = 0
	require.ErrorIs(t, migrateGachaEffort(state), ErrGachaRunState)
}
