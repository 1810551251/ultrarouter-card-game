package model

import (
	_ "embed"
	"fmt"

	"github.com/QuantumNous/new-api/common"
)

const gachaMasteryLootVersion = 6

type GachaCarryItem struct {
	Set    string             `json:"set"`
	Stat   string             `json:"stat"`
	Value  int                `json:"value"`
	Effect GachaArtifactBonus `json:"effect"`
	Recipe map[string]int     `json:"recipe"`
}

type GachaCarrySet struct {
	Name  string               `json:"name"`
	Tiers []GachaArtifactBonus `json:"tiers"`
}

type GachaCarryRules struct {
	Items map[string]GachaCarryItem `json:"items"`
	Sets  map[string]GachaCarrySet  `json:"sets"`
	Drops map[string]int            `json:"drops"`
}

//go:embed gacha_carry.json
var gachaCarryJSON []byte

var gachaCarry = func() GachaCarryRules {
	var rules GachaCarryRules
	if err := common.Unmarshal(gachaCarryJSON, &rules); err != nil {
		panic(err)
	}
	return rules
}()

func validGachaCarryState(state *GachaGameState) bool {
	worn := map[string]int{}
	for _, own := range state.Girls {
		if own == nil {
			return false
		}
		if own.Carry != "" {
			if _, ok := gachaCarry.Items[own.Carry]; !ok || own.Star == 0 {
				return false
			}
			worn[own.Carry]++
		}
	}
	for id, count := range state.CarryInventory {
		if _, ok := gachaCarry.Items[id]; !ok || count < 0 || count > gachaMaxResource-worn[id] {
			return false
		}
	}
	return true
}

func validateGachaCarryIntent(intent GachaActionIntent) error {
	if intent.Count != nil || intent.Enabled != nil || intent.Choice != nil || intent.Squad != nil || intent.Config != nil || intent.QueueID != nil || intent.RoomID != nil || intent.OpponentID != nil || intent.StakeE != nil {
		return ErrGachaIntent
	}
	switch intent.Action {
	case "carry_craft":
		if intent.GirlID != nil || intent.ItemID == nil {
			return ErrGachaIntent
		}
	case "carry_equip", "carry_remove":
		if intent.GirlID == nil || gachaGirl(*intent.GirlID) == nil || (intent.Action == "carry_equip") != (intent.ItemID != nil) {
			return ErrGachaIntent
		}
	default:
		return ErrGachaIntent
	}
	if intent.ItemID != nil {
		if _, ok := gachaCarry.Items[*intent.ItemID]; !ok {
			return ErrGachaIntent
		}
	}
	return nil
}

func applyGachaCarryAction(state *GachaGameState, intent GachaActionIntent, result *GachaActionResult) error {
	if state.CarryInventory == nil {
		state.CarryInventory = map[string]int{}
	}
	if intent.Action == "carry_craft" {
		id := *intent.ItemID
		item := gachaCarry.Items[id]
		owned := state.CarryInventory[id]
		if owned >= gachaMaxResource {
			return ErrGachaCapacity
		}
		for _, girl := range state.Girls {
			if girl.Carry == id {
				owned++
				if owned >= gachaMaxResource {
					return ErrGachaCapacity
				}
			}
		}
		for mat, cost := range item.Recipe {
			if state.Mats[mat] < cost {
				return ErrGachaFunds
			}
		}
		for mat, cost := range item.Recipe {
			state.Mats[mat] -= cost
		}
		state.CarryInventory[id]++
		result.Drops = append(result.Drops, gachaItems.Items[id].Name+" x1")
		return nil
	}
	own := state.Girls[*intent.GirlID]
	if !gachaCardOwned(own) {
		return ErrGachaIntent
	}
	next := ""
	if intent.ItemID != nil {
		next = *intent.ItemID
	}
	if own.Carry == next {
		return ErrGachaIntent
	}
	if next != "" && state.CarryInventory[next] < 1 {
		return ErrGachaFunds
	}
	if own.Carry != "" {
		if state.CarryInventory[own.Carry] >= gachaMaxResource {
			return ErrGachaCapacity
		}
		state.CarryInventory[own.Carry]++
		result.Drops = append(result.Drops, "已归还 · "+gachaItems.Items[own.Carry].Name)
	}
	if next != "" {
		state.CarryInventory[next]--
		result.Drops = append(result.Drops, "已携带 · "+gachaItems.Items[next].Name)
	}
	own.Carry = next
	return nil
}

func applyGachaCarryEffects(unit *gachaCombatUnit, state *GachaGameState, item GachaCarryItem) {
	applyGachaCultivationEffect(unit, item.Effect.Stat, item.Effect.Value)
	distinct := map[string]bool{}
	for _, id := range state.Squad {
		if own := state.Girls[id]; own != nil && own.Star > 0 && gachaCarry.Items[own.Carry].Set == item.Set {
			distinct[own.Carry] = true
		}
	}
	for _, tier := range gachaCarry.Sets[item.Set].Tiers {
		if len(distinct) >= tier.Count {
			applyGachaCultivationEffect(unit, tier.Stat, tier.Value)
		}
	}
}

func gachaMaterialStageKind(stage *GachaStage) string {
	switch {
	case stage.RelicTier >= 2:
		return "nightmare"
	case stage.RelicTier == 1:
		return "rift"
	case stage.HasBoss:
		return "campaign_boss"
	default:
		return "campaign"
	}
}

func awardGachaMasteryLoot(state *GachaGameState, stage *GachaStage, result *GachaFinishResult) error {
	amount := gachaCarry.Drops[gachaMaterialStageKind(stage)]
	if state.Mats["route_badge"] > gachaMaxResource-amount {
		return ErrGachaCapacity
	}
	if err := awardGachaTrainingLoot(state, stage, result); err != nil {
		return err
	}
	state.Mats["route_badge"] += amount
	result.Drops = append(result.Drops, fmt.Sprintf("%s x%d", gachaItems.Items["route_badge"].Name, amount))
	return nil
}
