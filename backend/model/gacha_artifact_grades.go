package model

import (
	_ "embed"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

type GachaArtifactQuality struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

type GachaArtifactFamily struct {
	Base       []int                   `json:"base"`
	RankEffect GachaRefinementEffect   `json:"rank_effect"`
	Rare       []GachaRefinementEffect `json:"rare"`
}

type GachaArtifactLevelCost struct {
	Base          int `json:"base"`
	Linear        int `json:"linear"`
	SquareDivisor int `json:"square_divisor"`
	OreBase       int `json:"ore_base"`
	OreEvery      int `json:"ore_every"`
}

type GachaArtifactGrowth struct {
	EarlySteps int `json:"early_steps"`
	LateEvery  int `json:"late_every"`
	LateHP     int `json:"late_hp"`
}

type GachaArtifactGradeRules struct {
	LevelMax    int                            `json:"level_max"`
	RankMax     int                            `json:"rank_max"`
	GemDefault  string                         `json:"gem_default"`
	ChipDefault string                         `json:"chip_default"`
	Qualities   []GachaArtifactQuality         `json:"qualities"`
	LevelCost   GachaArtifactLevelCost         `json:"level_cost"`
	RankLevels  []int                          `json:"rank_levels"`
	RankRivets  []int                          `json:"rank_rivets"`
	RankTickets []int                          `json:"rank_tickets"`
	Growth      GachaArtifactGrowth            `json:"growth"`
	Families    map[string]GachaArtifactFamily `json:"families"`
}

type GachaArtifactLoot struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

//go:embed gacha_artifact_grades.json
var gachaArtifactGradeJSON []byte

var gachaArtifactGrades = func() GachaArtifactGradeRules {
	var rules GachaArtifactGradeRules
	if err := common.Unmarshal(gachaArtifactGradeJSON, &rules); err != nil {
		panic(err)
	}
	ids := []string{"white", "green", "blue", "purple", "gold", "orange", "red"}
	if len(rules.Qualities) != len(ids) || rules.LevelMax != 120 || rules.RankMax != 12 || rules.GemDefault != "blue" || rules.ChipDefault != "purple" || rules.Growth.EarlySteps != 12 || rules.Growth.LateEvery < 1 || rules.Growth.LateHP < 1 || rules.LevelCost.Base < 1 || rules.LevelCost.Linear < 1 || rules.LevelCost.SquareDivisor < 1 || rules.LevelCost.OreBase < 1 || rules.LevelCost.OreEvery < 1 {
		panic("invalid artifact progression")
	}
	total := 0
	for index, quality := range rules.Qualities {
		if quality.ID != ids[index] || quality.Weight <= 0 || quality.Name == "" {
			panic("invalid artifact quality")
		}
		total += quality.Weight
	}
	if total != 10000 {
		panic("invalid artifact pool weights")
	}
	for _, costs := range [][]int{rules.RankLevels, rules.RankRivets, rules.RankTickets} {
		if len(costs) != rules.RankMax-1 || !slices.IsSorted(costs) || costs[0] <= 0 {
			panic("invalid artifact rank costs")
		}
	}
	knownEffects := []string{"rage", "rage_gain", "shield", "skill", "heal", "counter", "leech"}
	for _, family := range []string{"ruby", "jade", "sapphire", "wind", "amethyst", "edge", "aegis", "tempo"} {
		spec := rules.Families[family]
		if len(spec.Base) != len(ids) || spec.Base[0] < 0 || len(spec.Rare) != 3 || spec.RankEffect.Name == "" || !slices.Contains(knownEffects, spec.RankEffect.Effect) {
			panic("invalid artifact family")
		}
		for index := 1; index < len(spec.Base); index++ {
			if spec.Base[index] <= spec.Base[index-1] {
				panic("invalid graded base stats")
			}
		}
		for _, effect := range spec.Rare {
			if effect.Value < 1 || effect.Name == "" || !slices.Contains(knownEffects, effect.Effect) {
				panic("invalid rare artifact effect")
			}
		}
	}
	return rules
}()

type gachaArtifactKey struct {
	Base, Family, Quality string
	Level, Rank           int
	Chip                  bool
}

func gachaArtifactTemplate(base string) (gachaArtifactKey, bool) {
	if _, ok := gachaGemSpecs[base]; ok {
		return gachaArtifactKey{Base: base, Family: base, Quality: gachaArtifactGrades.GemDefault, Level: 1, Rank: 1}, true
	}
	family, pos, separated := strings.Cut(base, "_")
	_, known := gachaCultivation.Artifacts.ChipSets[family]
	if !separated || !known || len(pos) != 1 || pos[0] < '1' || int(pos[0]-'0') > len(gachaCultivation.Artifacts.ChipStars) {
		return gachaArtifactKey{}, false
	}
	return gachaArtifactKey{Base: base, Family: family, Quality: gachaArtifactGrades.ChipDefault, Level: 1, Rank: 1, Chip: true}, true
}

func (item gachaArtifactKey) defaultQuality() string {
	if item.Chip {
		return gachaArtifactGrades.ChipDefault
	}
	return gachaArtifactGrades.GemDefault
}

func (item gachaArtifactKey) id() string {
	id := item.Base
	if item.Quality != item.defaultQuality() {
		id += "~" + item.Quality
	}
	if item.Level == 1 && item.Rank == 1 {
		return id
	}
	legacy := item.Level - 1
	if item.Quality == item.defaultQuality() && legacy >= 1 && legacy <= 12 && item.Rank == max(1, legacy) {
		return id + "@" + strconv.Itoa(legacy)
	}
	return fmt.Sprintf("%s!%d.%d", id, item.Level, item.Rank)
}

func gachaArtifactGradeIndex(quality string) int {
	return slices.IndexFunc(gachaArtifactGrades.Qualities, func(grade GachaArtifactQuality) bool { return grade.ID == quality })
}

func parseGachaArtifactKey(id string) (gachaArtifactKey, bool) {
	if len(id) == 0 || len(id) > 40 {
		return gachaArtifactKey{}, false
	}
	head, progress := id, ""
	if index := strings.IndexAny(id, "@!"); index >= 0 {
		head, progress = id[:index], id[index:]
	}
	base, quality, graded := strings.Cut(head, "~")
	item, ok := gachaArtifactTemplate(base)
	if !ok {
		return gachaArtifactKey{}, false
	}
	if graded {
		if gachaArtifactGradeIndex(quality) < 0 || quality == item.Quality {
			return gachaArtifactKey{}, false
		}
		item.Quality = quality
	}
	if progress != "" {
		var err error
		if progress[0] == '@' {
			level, err := strconv.Atoi(progress[1:])
			if err != nil || graded || level < 1 || level > 12 {
				return gachaArtifactKey{}, false
			}
			item.Level, item.Rank = level+1, max(1, level)
		} else {
			level, rank, both := strings.Cut(progress[1:], ".")
			if !both {
				return gachaArtifactKey{}, false
			}
			item.Level, err = strconv.Atoi(level)
			if err != nil {
				return gachaArtifactKey{}, false
			}
			item.Rank, err = strconv.Atoi(rank)
			if err != nil {
				return gachaArtifactKey{}, false
			}
		}
	}
	if item.Level < 1 || item.Level > gachaArtifactGrades.LevelMax || item.Rank < 1 || item.Rank > gachaArtifactGrades.RankMax {
		return gachaArtifactKey{}, false
	}
	if item.Rank > 1 && item.Level < gachaArtifactGrades.RankLevels[item.Rank-2] && (graded || item.Rank > min(12, item.Level-1)) {
		return gachaArtifactKey{}, false
	}
	return item, item.id() == id
}

func gachaArtifactQualityAt(roll int) string {
	if roll < 0 {
		return ""
	}
	for _, grade := range gachaArtifactGrades.Qualities {
		if roll < grade.Weight {
			return grade.ID
		}
		roll -= grade.Weight
	}
	return ""
}

func (item gachaArtifactKey) name() string {
	return gachaArtifactGrades.Qualities[gachaArtifactGradeIndex(item.Quality)].Name + "·" + gachaItems.Items[item.Base].Name
}

func gachaArtifactLevelPrice(level int) (int, int) {
	step, cost := level-1, gachaArtifactGrades.LevelCost
	return cost.Base + cost.Linear*step + step*step/cost.SquareDivisor, cost.OreBase + step/cost.OreEvery
}

func advanceGachaArtifact(state *GachaGameState, intent GachaActionIntent, result *GachaActionResult) error {
	own := state.Girls[*intent.GirlID]
	item, ok := parseGachaArtifactKey(*intent.ItemID)
	if !ok || own == nil || own.Star < 1 {
		return ErrGachaIntent
	}
	index := *intent.Count - 1
	slots, mat, ticket := own.Gems, gachaEquipment.Refinement.GemMat, "gem_ticket"
	if item.Chip {
		slots, mat, ticket = own.Chips, gachaEquipment.Refinement.ChipMat, "chip_ticket"
		if index < 0 || index >= len(gachaCultivation.Artifacts.ChipStars) || own.Star < gachaCultivation.Artifacts.ChipStars[index] || own.ChipLocks&(1<<index) != 0 {
			return ErrGachaCultivationLocked
		}
	} else if index >= own.Sockets {
		return ErrGachaCultivationLocked
	}
	if index < 0 || index >= len(slots) || slots[index] != *intent.ItemID {
		return ErrGachaCultivationLocked
	}
	second, cost, extra := "forge_ore", 0, 0
	if strings.HasSuffix(intent.Action, "_rank") {
		if item.Rank >= gachaArtifactGrades.RankMax || item.Level < gachaArtifactGrades.RankLevels[item.Rank-1] {
			return ErrGachaCultivationLocked
		}
		mat, second, cost, extra = gachaEquipment.StarMat, ticket, gachaArtifactGrades.RankRivets[item.Rank-1], gachaArtifactGrades.RankTickets[item.Rank-1]
		item.Rank++
	} else {
		if item.Level >= gachaArtifactGrades.LevelMax {
			return ErrGachaCultivationLocked
		}
		cost, extra = gachaArtifactLevelPrice(item.Level)
		item.Level++
	}
	if state.Mats[mat] < cost || state.Mats[second] < extra {
		return ErrGachaFunds
	}
	state.Mats[mat] -= cost
	state.Mats[second] -= extra
	slots[index] = item.id()
	result.Drops = append(result.Drops, fmt.Sprintf("%s · Lv.%d / %d阶 · %s -%d · %s -%d", item.name(), item.Level, item.Rank, gachaItems.Items[mat].Name, cost, gachaItems.Items[second].Name, extra))
	return nil
}

func gachaEquippedArtifactGrades(own *GachaOwnedGirl) (GachaStats, []GachaRefinementEffect) {
	var flat GachaStats
	strongest := map[string]GachaRefinementEffect{}
	for _, slots := range [][]string{own.Gems, own.Chips} {
		for _, key := range slots {
			item, ok := parseGachaArtifactKey(key)
			if !ok {
				continue
			}
			spec := gachaEquipment.Refinement.Gems[item.Family]
			if item.Chip {
				spec = gachaEquipment.Refinement.Chips[item.Family]
			}
			grade := gachaArtifactGradeIndex(item.Quality)
			family, growth := gachaArtifactGrades.Families[item.Family], gachaArtifactGrades.Growth
			early := min(item.Level-1, growth.EarlySteps)
			late := item.Level - 1 - early
			value := family.Base[grade] + (early/spec.Every+late/(growth.LateEvery*spec.Every))*spec.Step
			if item.Base == "wind" {
				value -= 3
			}
			flat.addCultivationStat(spec.Stat, value)
			flat.HP += late * growth.LateHP
			keep := func(effect GachaRefinementEffect) {
				id := item.Family + ":" + effect.Name
				if effect.Value > strongest[id].Value {
					strongest[id] = effect
				}
			}
			primary := family.RankEffect
			primary.Value = item.Rank - 1
			keep(primary)
			for _, effect := range spec.Effects {
				if item.Rank >= effect.Level {
					keep(effect)
				}
			}
			for index, effect := range family.Rare {
				if grade >= index+4 {
					keep(effect)
				}
			}
		}
	}
	effects := make([]GachaRefinementEffect, 0, len(strongest))
	for _, id := range slices.Sorted(maps.Keys(strongest)) {
		effects = append(effects, strongest[id])
	}
	return flat, effects
}
