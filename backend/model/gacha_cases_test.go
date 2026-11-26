package model

import (
	"maps"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCaseCommand(action, id, choice string, count int) GachaActionIntent {
	intent := GachaActionIntent{RequestID: uuid.NewString(), Action: action, ItemID: &id}
	if action != "case_open" {
		intent.Count = &count
		intent.Choice = &choice
	}
	return intent
}

func TestGachaCasesPathsBudgetsAndExpiry(t *testing.T) {
	for _, sample := range []struct {
		choices       []string
		close, ending string
		total         int
	}{{nil, "withdraw", "withdrawn", 0}, {[]string{"survey"}, "file", "filed", 2}, {[]string{"survey", "survey", "survey"}, "pursue", "revealed", 6}, {[]string{"veil", "veil", "veil"}, "pursue", "partial", 2}, {[]string{"dare", "dare"}, "", "overrun", 0}, {[]string{"dare", "veil", "dare"}, "pursue", "overrun", 0}} {
		state := &GachaGameState{Daily: GachaGameDaily{Date: "2026-11-26"}}
		out := &GachaActionResult{}
		open := testCaseCommand("case_open", "silent_bell", "", 0)
		require.NoError(t, applyGachaCaseAction(state, open, out))
		id := state.Casebook.Active.ID
		for index, choice := range sample.choices {
			require.NoError(t, applyGachaCaseAction(state, testCaseCommand("case_step", id, choice, index+1), out))
			require.NoError(t, normalizeGachaCasebook(state))
		}
		if sample.close != "" {
			require.NoError(t, applyGachaCaseAction(state, testCaseCommand("case_close", id, sample.close, len(sample.choices)), out))
		}
		assert.Nil(t, state.Casebook.Active)
		require.NotNil(t, state.Casebook.Last)
		assert.Equal(t, sample.ending, state.Casebook.Last.Ending)
		assert.Equal(t, sample.total, state.Casebook.Awarded)
		assert.Equal(t, 1, gachaCaseEndingCount(state, "silent_bell", sample.ending))
		require.Error(t, applyGachaCaseAction(state, testCaseCommand("case_close", id, "file", 1), out))
		require.Error(t, applyGachaCaseAction(state, open, out))
	}
	state := &GachaGameState{Daily: GachaGameDaily{Date: "2026-11-26"}}
	out := &GachaActionResult{}
	for _, file := range []string{"silent_bell", "glass_clock"} {
		open := testCaseCommand("case_open", file, "", 0)
		require.NoError(t, applyGachaCaseAction(state, open, out))
		for index := 1; index <= 3; index++ {
			require.NoError(t, applyGachaCaseAction(state, testCaseCommand("case_step", open.RequestID, "survey", index), out))
		}
		require.NoError(t, applyGachaCaseAction(state, testCaseCommand("case_close", open.RequestID, "pursue", 3), out))
	}
	assert.Equal(t, 10, state.Casebook.Awarded)
	require.ErrorIs(t, applyGachaCaseAction(state, testCaseCommand("case_open", "salt_lantern", "", 0), out), ErrGachaCultivationLocked)
	evidence := maps.Clone(state.Casebook.Evidence)
	state.Daily.Date = "2026-11-27"
	require.NoError(t, normalizeGachaCasebook(state))
	require.Empty(t, state.Casebook.Opened)
	assert.Zero(t, state.Casebook.Awarded)
	assert.Equal(t, evidence, state.Casebook.Evidence)
	open := testCaseCommand("case_open", "salt_lantern", "", 0)
	require.NoError(t, applyGachaCaseAction(state, open, out))
	require.NoError(t, applyGachaCaseAction(state, testCaseCommand("case_step", open.RequestID, "survey", 1), out))
	state.Daily.Date = "2026-11-28"
	require.NoError(t, normalizeGachaCasebook(state))
	assert.Nil(t, state.Casebook.Active)
	assert.Equal(t, "withdrawn", state.Casebook.Last.Ending)
	assert.Equal(t, evidence, state.Casebook.Evidence)
}

func TestGachaCaseFrozenOathsStaleChoicesAndAtomicCapacity(t *testing.T) {
	for _, path := range []string{"lantern", "tide", "ward"} {
		leader := "c01"
		state := battleTestState([]string{leader}, 1, 1)
		state.Daily.Date = "2026-11-26"
		state.Girls[leader].Vigil = &GachaVigil{Path: path, Rank: 1}
		out := &GachaActionResult{}
		open := testCaseCommand("case_open", "silent_bell", "", 0)
		open.GirlID = &leader
		require.NoError(t, applyGachaCaseAction(state, open, out))
		state.Girls[leader].Vigil = nil
		first := "dare"
		if path == "lantern" {
			first = "survey"
		}
		require.NoError(t, applyGachaCaseAction(state, testCaseCommand("case_step", open.RequestID, first, 1), out))
		assert.Equal(t, path, state.Casebook.Active.Oath)
		require.Error(t, applyGachaCaseAction(state, testCaseCommand("case_step", open.RequestID, first, 1), out))
		require.Error(t, applyGachaCaseAction(state, testCaseCommand("case_close", uuid.NewString(), "file", 1), out))
		if path == "lantern" {
			assert.Equal(t, 3, state.Casebook.Active.Clues["echo"])
			state.Casebook.Evidence["echo"] = gachaMaxResource
			require.ErrorIs(t, applyGachaCaseAction(state, testCaseCommand("case_close", open.RequestID, "file", 1), out), ErrGachaCapacity)
			assert.NotNil(t, state.Casebook.Active)
			assert.Zero(t, state.Casebook.Awarded)
		}
		if path == "tide" {
			assert.Equal(t, 2, state.Casebook.Active.Pressure)
		}
		if path == "ward" {
			require.NoError(t, applyGachaCaseAction(state, testCaseCommand("case_step", open.RequestID, "dare", 2), out))
			assert.Equal(t, 5, state.Casebook.Active.Pressure)
		}
		assert.True(t, state.Casebook.Active.OathUsed)
		require.NoError(t, normalizeGachaCasebook(state))
		state.Casebook.Active.Pressure++
		require.ErrorIs(t, normalizeGachaCasebook(state), ErrGachaRunState)
	}
}

func TestGachaCaseMalformedActiveAndPayloads(t *testing.T) {
	state := &GachaGameState{Daily: GachaGameDaily{Date: "2026-11-28"}}
	open := testCaseCommand("case_open", "silent_bell", "", 0)
	require.NoError(t, applyGachaCaseAction(state, open, &GachaActionResult{}))
	state.Casebook.Active.Clues = nil
	require.ErrorIs(t, normalizeGachaCasebook(state), ErrGachaRunState)
	for _, intent := range []GachaActionIntent{testCaseCommand("case_open", "unknown", "", 0), testCaseCommand("case_step", open.RequestID, "survey", 0), testCaseCommand("case_close", open.RequestID, "survey", 2), testCaseCommand("case_step", "silent_bell", "veil", 1)} {
		assert.ErrorIs(t, ValidateGachaAction(intent), ErrGachaIntent)
	}
	open.Materials = map[string]int{"ember": 400}
	assert.ErrorIs(t, ValidateGachaAction(open), ErrGachaIntent)
}
