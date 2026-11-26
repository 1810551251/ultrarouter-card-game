package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVigilBranchesFormationAndOneShotCombat(t *testing.T) {
	ids := []string{"c01", "c08", "c06", "c12"}
	state := battleTestState(ids, 40, 3)
	for index, choice := range [][2]string{{"lantern", "beacon"}, {"lantern", "flare"}, {"tide", "undertow"}, {"ward", "mirror"}} {
		state.Girls[ids[index]].Vigil = &GachaVigil{Path: choice[0], Aspect: choice[1], Rank: 2, Marks: 9, Overdraw: index == 0}
	}
	for _, side := range []string{"ally", "enemy"} {
		state.previousTraining = true
		old, err := gachaSquadUnits(state, side, "f")
		require.NoError(t, err)
		state.previousTraining = false
		units, err := gachaSquadUnits(state, side, "f")
		require.NoError(t, err)
		assert.Equal(t, old[0].MaxHP-old[0].MaxHP*12/100, units[0].HP)
		assert.Equal(t, old[0].Rage+8, units[0].Rage)
		assert.Equal(t, old[0].Shield+units[0].MaxHP*5/100, units[0].Shield)
		assert.InDelta(t, old[0].skillBoost+0.12, units[0].skillBoost, 0.000001)
		enemy := old[0]
		enemy.Rage = 30
		combat := &gachaCombat{allies: units, waves: [][]*gachaCombatUnit{{enemy}}, report: &GachaBattleReport{}, formula: gachaCatalog.Formula}
		combat.invokeVigil(units[0], units, []*gachaCombatUnit{enemy}, "cast")
		assert.Equal(t, units[0].MaxHP, units[0].HP)
		combat.invokeVigil(units[0], units, []*gachaCombatUnit{enemy}, "cast")
		require.Len(t, combat.report.Events, 1)
		combat.invokeVigil(units[2], units, []*gachaCombatUnit{enemy}, "cast")
		assert.Equal(t, 23, enemy.Rage)
		combat.invokeVigil(units[3], units, nil, "ready")
		assert.InDelta(t, old[3].counter+0.07, units[3].counter, 0.000001)
	}
}

func TestVigilSettlementWitnessAndRitualCosts(t *testing.T) {
	state := battleTestState([]string{"c01", "c08"}, 20, 2)
	apply := func(action, choice string) error {
		return applyGachaVigil(state, trainingTestIntent(action, choice), &GachaActionResult{})
	}
	require.NoError(t, apply("vigil_bind", "lantern"))
	require.ErrorIs(t, apply("vigil_bind", "tide"), ErrGachaCultivationLocked)
	state.Mats["break_seal"] = 17
	own := state.Girls["c01"]
	snap := GachaRunSnapshot{Squad: state.Squad}
	issueGachaVigil(state, &snap)
	state.Girls["c08"].Vigil = &GachaVigil{Path: "ward"}
	report := &GachaBattleReport{AliveN: 2, TotalN: 2}
	result := &GachaFinishResult{Win: true}
	for range 4 {
		settleGachaVigil(state, &GachaStage{}, &snap, report, result)
	}
	assert.Equal(t, 3, own.Vigil.Marks)
	assert.Zero(t, state.Girls["c08"].Vigil.Marks)
	require.NoError(t, apply("vigil_ascend", ""))
	assert.Equal(t, 15, state.Mats["break_seal"])
	own.Vigil.Marks = 9
	require.ErrorIs(t, apply("vigil_ascend", "anchor"), ErrGachaCultivationLocked)
	require.ErrorIs(t, apply("vigil_ascend", "beacon"), ErrGachaFunds)
	assert.Equal(t, 15, state.Mats["break_seal"])
	stock := gachaEvidence(state)
	stock["ember"], stock["echo"], stock["key"] = 4, 3, 3
	require.NoError(t, apply("vigil_ascend", "beacon"))
	assert.Equal(t, 2, stock["ember"])
	require.NoError(t, apply("vigil_refocus", "flare"))
	assert.Equal(t, "flare", own.Vigil.Aspect)
	stance := trainingTestIntent("vigil_stance", "")
	enable := true
	stance.Enabled = &enable
	require.NoError(t, applyGachaVigil(state, stance, &GachaActionResult{}))
	issueGachaVigil(state, &snap)
	for range 3 {
		settleGachaVigil(state, &GachaStage{}, &snap, report, result)
	}
	assert.Equal(t, 6, own.Vigil.Strain)
	assert.False(t, own.Vigil.Overdraw)
	require.ErrorIs(t, applyGachaVigil(state, stance, &GachaActionResult{}), ErrGachaCultivationLocked)
	require.NoError(t, apply("vigil_ground", ""))
	assert.Equal(t, 3, own.Vigil.Strain)
	issueGachaVigil(state, &snap)
	settleGachaVigil(state, &GachaStage{}, &snap, report, result)
	assert.Equal(t, 2, own.Vigil.Strain)
	require.NoError(t, applyGachaVigil(state, stance, &GachaActionResult{}))
}
