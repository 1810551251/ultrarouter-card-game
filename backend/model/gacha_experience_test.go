package model

import (
	"fmt"
	"testing"
	"time"

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
	var drops []string
	require.NoError(t, awardGachaExperience(state, state.Squad, 0, &drops))
	require.Zero(t, own.Level)
	require.NoError(t, awardGachaExperience(state, state.Squad, 99, &drops))
	require.Zero(t, own.Level)
	require.NoError(t, awardGachaExperience(state, state.Squad, 1, &drops))
	require.Equal(t, 1, own.Level)
	require.Zero(t, own.Exp)
	assert.False(t, gachaCardOwned(&GachaOwnedGirl{Star: 1, Level: 0}))
	state.EffortVersion = 0
	require.ErrorIs(t, migrateGachaEffort(state), ErrGachaRunState)
}

func TestGachaExperienceCreditsOnlyWhatEachPartnerCanKeep(t *testing.T) {
	state := battleTestState([]string{"c08", "c06", "c01"}, 1, 1)
	state.Captain = &GachaCaptain{Level: 3}
	capped, partial, full := state.Girls["c08"], state.Girls["c06"], state.Girls["c01"]
	capped.Level, capped.Exp = 3, 73
	partial.Level, partial.Exp = 2, gachaLevelCost(2)-5
	var drops []string
	require.NoError(t, awardGachaExperience(state, state.Squad, 40, &drops))
	require.Equal(t, []string{fmt.Sprintf("伙伴经验：%s +0（已达等级上限）、%s +5（已达等级上限）、%s +40", gachaGirl("c08").Name, gachaGirl("c06").Name, gachaGirl("c01").Name)}, drops)
	assert.Equal(t, 73, capped.Exp)
	assert.Equal(t, 3, partial.Level)
	assert.Zero(t, partial.Exp)
	assert.Equal(t, 57, full.Exp)
	state.Captain.Level = gachaCatalog.Growth.MaxLevel
	for _, own := range state.Girls {
		own.Level = state.Captain.Level
	}
	drops = nil
	require.NoError(t, awardGachaExperience(state, state.Squad, 40, &drops))
	assert.Equal(t, []string{fmt.Sprintf("伙伴经验 +0（全员已达 Lv.%d 上限）", state.Captain.Level)}, drops)
}

func TestGachaExperienceReceiptsAcrossBattleSettlements(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, gameZone)
	for _, stageID := range []string{"s1", "r1", "training", "expedition", "res_manifest", "v1h1"} {
		t.Run(stageID, func(t *testing.T) {
			state := newGachaState(now)
			state.Captain.Level = 2
			state.Squad = state.Squad[:1]
			partner := state.Squad[0]
			state.Girls[partner].Exp = gachaLevelCost(1) - 3
			snapshot := GachaRunSnapshot{Squad: state.Squad, RulesVersion: gachaStaminaXPVersion}
			report, result := &GachaBattleReport{Win: true, Stars: 2}, &GachaFinishResult{Win: true}
			switch stageID {
			case "training":
				state.ProgressionRules = DefaultGachaProgressionConfig()
				require.NoError(t, normalizeGachaProgression(state, now, state.ProgressionRules))
				_, issued, err := specialGachaProgressionStage(stageID, state, now)
				require.NoError(t, err)
				require.NoError(t, settleGachaProgressionMode(state, &issued, result, now))
				assert.Contains(t, result.Drops, "训练不产出 E")
			case "expedition":
				resetGachaWeeklyModes(state, now)
				state.Expedition.Route = "assault"
				stage, issued, err := specialGachaStage(stageID, state, now)
				require.NoError(t, err)
				issued.Squad = state.Squad
				require.NoError(t, settleGachaMode(state, stage, &issued, report, result, false))
			case "res_manifest", "v1h1":
				require.NoError(t, settleGachaVoyagePoint(state, gachaVoyage.Stages[stageID], &snapshot, report, result))
			default:
				require.NoError(t, awardGachaGameRewards(state, gachaStage(stageID), &snapshot, report, result))
			}
			assert.Contains(t, result.Drops, "伙伴经验："+gachaGirl(partner).Name+" +3（已达等级上限）")
			assert.Equal(t, 2, state.Girls[partner].Level)
			assert.Zero(t, state.Girls[partner].Exp)
		})
	}
}
