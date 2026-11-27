package model

import (
	"maps"
	"slices"
	"strings"
)

type GachaRefinementEffect struct {
	Level  int    `json:"level"`
	Name   string `json:"name"`
	Effect string `json:"effect"`
	Value  int    `json:"value"`
}

type GachaRefinementItem struct {
	Stat    string                  `json:"stat"`
	Base    int                     `json:"base"`
	Step    int                     `json:"step"`
	Every   int                     `json:"every"`
	Effects []GachaRefinementEffect `json:"effects"`
}

type GachaRefinementRules struct {
	Costs    []int                          `json:"costs"`
	OreCosts []int                          `json:"ore_costs"`
	GemMat   string                         `json:"gem_mat"`
	ChipMat  string                         `json:"chip_mat"`
	Gems     map[string]GachaRefinementItem `json:"gems"`
	Chips    map[string]GachaRefinementItem `json:"chips"`
}

func validGachaRefinementRules(rules GachaRefinementRules) bool {
	if len(rules.Costs) == 0 || len(rules.Costs) != len(rules.OreCosts) || rules.GemMat == "" || rules.ChipMat == "" {
		return false
	}
	for index, cost := range rules.Costs {
		if cost <= 0 || rules.OreCosts[index] <= 0 {
			return false
		}
	}
	for _, items := range []map[string]GachaRefinementItem{rules.Gems, rules.Chips} {
		for _, item := range items {
			if item.Every < 1 || item.Base < 1 || item.Step < 1 {
				return false
			}
			for _, effect := range item.Effects {
				if effect.Level < 1 || effect.Level > len(rules.Costs) || effect.Value <= 0 {
					return false
				}
			}
		}
	}
	return true
}

func splitGachaRefinement(id string) (string, int, bool) {
	item, ok := parseGachaArtifactKey(id)
	return item.Base, max(0, min(12, item.Level-1)), ok
}

func validateGachaRefinementIntent(intent GachaActionIntent) error {
	if intent.GirlID == nil || gachaGirl(*intent.GirlID) == nil || intent.Count == nil || *intent.Count < 1 || intent.ItemID == nil {
		return ErrGachaIntent
	}
	if strings.HasPrefix(intent.Action, "gem_") {
		if _, ok := gachaGem(*intent.ItemID); !ok || *intent.Count > len(gachaCultivation.Artifacts.GemSockets) {
			return ErrGachaIntent
		}
	} else if _, slot, ok := parseGachaChip(*intent.ItemID); !ok || slot != *intent.Count {
		return ErrGachaIntent
	}
	return nil
}

func gachaEquippedRefinement(own *GachaOwnedGirl) (GachaStats, []GachaRefinementEffect) {
	var flat GachaStats
	var effects []GachaRefinementEffect
	rules := &gachaEquipment.Refinement
	for index, slots := range [][]string{own.Gems, own.Chips} {
		highest := map[string]int{}
		items := rules.Gems
		if index == 1 {
			items = rules.Chips
		}
		for _, key := range slots {
			base, level, ok := splitGachaRefinement(key)
			if !ok {
				continue
			}
			if index == 1 {
				base, _, _ = strings.Cut(base, "_")
			}
			item, exists := items[base]
			if !exists {
				continue
			}
			flat.addCultivationStat(item.Stat, item.Base+level/item.Every*item.Step)
			highest[base] = max(highest[base], level)
		}
		for _, family := range slices.Sorted(maps.Keys(highest)) {
			for _, effect := range items[family].Effects {
				if highest[family] >= effect.Level {
					effects = append(effects, effect)
				}
			}
		}
	}
	return flat, effects
}
