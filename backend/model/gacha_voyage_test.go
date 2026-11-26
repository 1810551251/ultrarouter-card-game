package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVoyageRoutesHaveIndependentKeysAndRealBattleDefinitions(t *testing.T) {
	seen := map[string]bool{}
	chests := map[string]bool{}
	require.Len(t, gachaVoyage.Chapters, 4)
	for _, chapter := range gachaVoyage.Chapters {
		for mode, length := range map[string]int{"story": 8, "hard": 3, "nightmare": 2} {
			route := chapter.Routes[mode]
			require.Len(t, route, length)
			for _, id := range route {
				require.False(t, seen[id])
				seen[id] = true
				point := gachaVoyage.Stages[id]
				require.Equal(t, mode, point.Kind)
				require.Equal(t, &point.GachaStage, gachaStage(id))
				require.Positive(t, point.Cost)
				for _, prior := range append(append([]string{}, point.Requires...), point.ChapterGate) {
					if prior != "" {
						require.NotNil(t, gachaStage(prior))
					}
				}
			}
			for step, chest := range chapter.Chests[mode] {
				require.Equal(t, length*(step+1), chest.Stars)
				require.False(t, chests[chest.ID])
				chests[chest.ID] = true
				entry, ids := findGachaVoyageChest(chest.ID)
				require.Equal(t, chest, *entry)
				require.Equal(t, route, ids)
			}
		}
	}
	require.Len(t, seen, 52)
	require.Len(t, chests, 36)
	require.Len(t, gachaVoyage.Resources, 6)
	for _, id := range gachaVoyage.Resources {
		point := gachaVoyage.Stages[id]
		require.Equal(t, "resources", point.Kind)
		require.Equal(t, 3, point.DailyLimit)
		require.NotEmpty(t, point.Prize)
		_, walletRule := gachaStageRules[id]
		require.False(t, walletRule)
		for _, prize := range point.Prize {
			if prize.Item != "gold" {
				require.NotEmpty(t, gachaItems.Items[prize.Item].Name)
			}
		}
	}
}

func TestVoyageFreshOldChapterAndDifficultyGates(t *testing.T) {
	state := newGachaState(time.Now())
	state.Captain = &GachaCaptain{Level: 7}
	state.Cleared["s2"] = true
	second := gachaVoyage.Stages["s4"]
	require.ErrorIs(t, checkGachaVoyagePoint(state, second), ErrGachaLocked)
	state.Cleared[second.ChapterGate] = true
	require.NoError(t, checkGachaVoyagePoint(state, second))
	delete(state.Cleared, second.ChapterGate)
	state.VoyageEdition = 0
	migrateGachaVoyage(state)
	require.NoError(t, checkGachaVoyagePoint(state, second))
	assert.Zero(t, state.Stars[second.ChapterGate])
	state.Cleared["s6"] = true
	migrateGachaVoyage(state)
	assert.False(t, state.VoyagePasses["s7"], "migration must not open future chapters on later reads")
	hard := gachaVoyage.Stages["v1h1"]
	state.Cleared[hard.Requires[0]] = true
	require.ErrorIs(t, checkGachaVoyagePoint(state, hard), ErrGachaLocked)
	state.Captain.Level = 8
	require.NoError(t, checkGachaVoyagePoint(state, hard))
	state.Daily.VoyageAttempts = map[string]int{hard.ID: 3}
	require.ErrorIs(t, checkGachaVoyagePoint(state, hard), ErrGachaVoyageLimit)
	secondHard := gachaVoyage.Stages["v2h1"]
	state.Cleared[secondHard.Requires[0]] = true
	require.ErrorIs(t, checkGachaVoyagePoint(state, secondHard), ErrGachaLocked)
	state.Cleared[secondHard.Requires[1]] = true
	require.NoError(t, checkGachaVoyagePoint(state, secondHard))
}

func TestVoyageChestCannotBorrowStarsOrClaimTwice(t *testing.T) {
	state := newGachaState(time.Now())
	result := &GachaActionResult{}
	id := "c1-hard-1"
	state.Stars["s1"] = 3
	state.Stars["v2h1"] = 3
	state.Stars["v1h1"] = 2
	require.ErrorIs(t, claimGachaVoyageChest(state, id, result), ErrGachaVoyageChest)
	state.Stars["v1h1"] = 3
	state.Mats["training_manual"] = gachaMaxResource
	before := state.Gold
	require.ErrorIs(t, claimGachaVoyageChest(state, id, result), ErrGachaCapacity)
	assert.Equal(t, before, state.Gold)
	assert.False(t, state.VoyageClaims[id])
	state.Mats["training_manual"] = 0
	require.NoError(t, claimGachaVoyageChest(state, id, result))
	assert.True(t, state.VoyageClaims[id])
	assert.Len(t, result.Drops, 2)
	assert.ErrorIs(t, claimGachaVoyageChest(state, id, result), ErrGachaVoyageChest)
	intent := GachaActionIntent{Action: "voyage_chest", RequestID: uuid.NewString(), ItemID: &id}
	require.NoError(t, ValidateGachaAction(intent))
	count := 10
	intent.Count = &count
	assert.ErrorIs(t, ValidateGachaAction(intent), ErrGachaIntent)
}

func TestVoyageReplayKeepsBestStarsAndSeparatesOneTimeAwards(t *testing.T) {
	state := newGachaState(time.Now())
	state.Captain = &GachaCaptain{Level: 8}
	point := gachaVoyage.Stages["v1h1"]
	snapshot := &GachaRunSnapshot{Squad: state.Squad}
	for _, stars := range []int{1, 3, 2, 3} {
		result := &GachaFinishResult{Win: true}
		report := &GachaBattleReport{Win: true, Stars: stars}
		require.NoError(t, awardGachaVoyageMilestones(state, point, stars, result))
		require.NoError(t, settleGachaVoyagePoint(state, point, snapshot, report, result))
	}
	assert.Equal(t, 3, state.Stars[point.ID])
	assert.True(t, state.Cleared[point.ID])
	assert.Equal(t, 1, state.Mats["field_notes"])
	assert.Equal(t, 1, state.Mats["expedition_atlas"])
}
