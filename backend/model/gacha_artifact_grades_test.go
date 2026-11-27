package model

import (
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGachaArtifactGradesCanonicalAndWeightIntervals(t *testing.T) {
	start := 0
	for _, grade := range gachaArtifactGrades.Qualities {
		assert.Equal(t, grade.ID, gachaArtifactQualityAt(start))
		start += grade.Weight
		assert.Equal(t, grade.ID, gachaArtifactQualityAt(start-1))
		for _, base := range []string{"ruby", "jade", "wind", "sapphire", "amethyst", "edge_1", "tempo_6", "aegis_3"} {
			item, ok := gachaArtifactTemplate(base)
			require.True(t, ok)
			item.Quality = grade.ID
			for _, pair := range [][2]int{{1, 1}, {2, 1}, {2, 2}, {13, 1}, {110, 12}, {120, 12}} {
				item.Level, item.Rank = pair[0], pair[1]
				round, valid := parseGachaArtifactKey(item.id())
				require.True(t, valid, item.id())
				assert.Equal(t, item, round)
			}
		}
	}
	assert.Equal(t, 10000, start)
	assert.Empty(t, gachaArtifactQualityAt(-1))
	assert.Empty(t, gachaArtifactQualityAt(10000))
	for _, base := range []string{"ruby", "edge_1"} {
		for old := 1; old <= 12; old++ {
			item, ok := parseGachaArtifactKey(fmt.Sprintf("%s@%d", base, old))
			require.True(t, ok)
			assert.Equal(t, old+1, item.Level)
			assert.Equal(t, old, item.Rank)
		}
	}
	for _, key := range []string{"ruby!14.12", "edge_1!13.11", "ruby!2.2", "ruby!3.1"} {
		_, ok := parseGachaArtifactKey(key)
		assert.True(t, ok, key)
	}
	for _, key := range []string{"ruby~blue", "tempo_1~purple", "ruby~red@1", "ruby!1.1", "ruby!7.6", "ruby~red!13.12", "ruby~red!5.4", "ruby~red!120.13", "ruby~red!121.12", "ruby!12.12", "ruby!02.1", "edge_7~red", "fake~red", "ruby@+1", "ruby~~red", "ruby~red!2.2!3.1", "ruby!2.+1"} {
		_, ok := parseGachaArtifactKey(key)
		assert.False(t, ok, key)
	}
	flat, ore := gachaArtifactLevelPrice(119)
	assert.Equal(t, 378, flat)
	assert.Equal(t, 6, ore)
}

func TestGachaArtifactGradeCombatProcsAndFrozenV7(t *testing.T) {
	state := battleTestState([]string{"c01"}, 40, 1)
	own := state.Girls["c01"]
	own.Sockets = 3
	unit := func(snapshot *GachaGameState, side string) *gachaCombatUnit {
		units, err := gachaSquadUnits(snapshot, side, "grade")
		require.NoError(t, err)
		return units[0]
	}
	own.Gems = []string{"ruby@6"}
	previous, err := common.Marshal(gachaFrozenRoster{state, 7})
	require.NoError(t, err)
	frozen, err := thawGachaRoster(GachaDocument(previous))
	require.NoError(t, err)
	old, current := unit(frozen, "enemy"), unit(state, "enemy")
	assert.Equal(t, old.Atk, current.Atk)
	assert.InDelta(t, .04, old.skillBoost, 1e-9)
	assert.InDelta(t, .09, current.skillBoost, 1e-9)
	own.Gems = []string{"ruby~red!110.12", "ruby~gold!110.12", "ruby~white"}
	own.Chips = []string{"edge_1~red"}
	red := unit(state, "ally")
	enemy := unit(state, "enemy")
	assert.Equal(t, red.MaxHP, enemy.MaxHP)
	assert.Equal(t, red.Rage, enemy.Rage)
	assert.Equal(t, 12, red.Rage)
	assert.InDelta(t, .02, red.lifeSteal, 1e-9)
	assert.InDelta(t, .04, red.counter, 1e-9)
	assert.InDelta(t, .15, red.skillBoost, 1e-9)
	assert.Equal(t, 2+3, red.rageBonus)
	red.HP -= 100
	enemy.HP, enemy.MaxHP = 10000, 10000
	battle := gachaCombat{formula: gachaCatalog.Formula, rng: battleTestRNG()}
	hit := battle.damage(red, enemy, 1, 0)
	assert.Positive(t, hit.SelfHeal)
	assert.Positive(t, hit.Reflected)
	own.Gems = []string{"wind~white"}
	white := unit(state, "ally")
	own.Gems = []string{"wind~red"}
	highest := unit(state, "ally")
	assert.Equal(t, 7, highest.Spd-white.Spd)
	own.Gems = []string{"wind~red!119.1"}
	first := unit(state, "ally")
	own.Gems = []string{"wind~red!120.1"}
	last := unit(state, "ally")
	assert.Equal(t, first.MaxHP+2, last.MaxHP)
	state.GemInventory = map[string]int{"wind~red!120.12": 100}
	assert.Equal(t, last.GachaBattleUnit, unit(state, "ally").GachaBattleUnit)
}
