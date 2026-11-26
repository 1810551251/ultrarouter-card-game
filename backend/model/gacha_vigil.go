package model

import (
	_ "embed"
	"fmt"
	"slices"

	"github.com/QuantumNous/new-api/common"
)

type GachaVigilAspect struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Story       string         `json:"story"`
	Effect      string         `json:"effect"`
	EffectLabel string         `json:"effect_label"`
	Evidence    map[string]int `json:"evidence"`
}

type GachaVigilPath struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Motto       string             `json:"motto"`
	Task        string             `json:"task"`
	Effect      string             `json:"effect"`
	EffectLabel string             `json:"effect_label"`
	Aspects     []GachaVigilAspect `json:"aspects"`
}

type GachaVigilRank struct {
	Name     string         `json:"name"`
	Marks    int            `json:"marks"`
	Seals    int            `json:"seals"`
	Effect   int            `json:"effect"`
	Evidence map[string]int `json:"evidence,omitempty"`
}

type GachaVigilRules struct {
	Paths          []GachaVigilPath `json:"paths"`
	Ranks          []GachaVigilRank `json:"ranks"`
	DailyLimit     int              `json:"daily_limit"`
	OmenBonus      int              `json:"omen_bonus"`
	OmenHPCost     int              `json:"omen_hp_cost"`
	StrainMax      int              `json:"strain_max"`
	StrainStartMax int              `json:"strain_start_max"`
	StrainGain     int              `json:"strain_gain"`
	GroundDrop     int              `json:"ground_drop"`
	GroundEvidence string           `json:"ground_evidence"`
	RefocusSeals   int              `json:"refocus_seals"`
	ChorusRage     int              `json:"chorus_rage"`
	TwinsShield    int              `json:"twins_shield"`
}

type GachaVigil struct {
	Path     string `json:"path"`
	Rank     int    `json:"rank"`
	Marks    int    `json:"marks"`
	Aspect   string `json:"aspect,omitempty"`
	Strain   int    `json:"strain"`
	Overdraw bool   `json:"overdraw"`
}

//go:embed gacha_vigil.json
var gachaVigilJSON []byte

var gachaVigil = func() GachaVigilRules {
	var rules GachaVigilRules
	if err := common.Unmarshal(gachaVigilJSON, &rules); err != nil {
		panic(err)
	}
	return rules
}()

func vigilPath(id string) *GachaVigilPath {
	for index := range gachaVigil.Paths {
		if path := &gachaVigil.Paths[index]; path.ID == id {
			return path
		}
	}
	return nil
}

func gachaCaseOath(own *GachaOwnedGirl) string {
	if !gachaCardOwned(own) || own.Vigil == nil || own.Vigil.Rank < 1 {
		return ""
	}
	return own.Vigil.Path
}

func (path *GachaVigilPath) aspect(id string) *GachaVigilAspect {
	for index := range path.Aspects {
		if aspect := &path.Aspects[index]; aspect.ID == id {
			return aspect
		}
	}
	return nil
}

func validGachaVigil(own *GachaOwnedGirl) bool {
	v := own.Vigil
	if v == nil {
		return true
	}
	path := vigilPath(v.Path)
	return gachaCardOwned(own) && path != nil && v.Rank >= 0 && v.Rank <= len(gachaVigil.Ranks) && v.Marks >= 0 && v.Marks <= gachaVigil.Ranks[len(gachaVigil.Ranks)-1].Marks &&
		(v.Rank == 0 || v.Marks >= gachaVigil.Ranks[v.Rank-1].Marks) && (v.Rank < 2 && v.Aspect == "" || v.Rank >= 2 && path.aspect(v.Aspect) != nil) &&
		v.Strain >= 0 && v.Strain <= gachaVigil.StrainMax && (!v.Overdraw || v.Strain < gachaVigil.StrainMax && v.Rank > 0) && (v.Rank > 0 || v.Strain == 0)
}

func validateGachaVigilIntent(intent GachaActionIntent) error {
	if intent.GirlID == nil || gachaGirl(*intent.GirlID) == nil || intent.ItemID != nil || intent.Count != nil || intent.Squad != nil || intent.OpponentID != nil || intent.QueueID != nil || intent.RoomID != nil || intent.StakeE != nil || intent.Config != nil || intent.Materials != nil {
		return ErrGachaIntent
	}
	choice := intent.Choice
	if intent.Action == "vigil_stance" && intent.Enabled != nil && choice == nil {
		return nil
	}
	if intent.Enabled != nil {
		return ErrGachaIntent
	}
	switch intent.Action {
	case "vigil_bind":
		if choice != nil && vigilPath(*choice) != nil {
			return nil
		}
	case "vigil_ground":
		if choice == nil {
			return nil
		}
	case "vigil_ascend", "vigil_refocus":
		if choice == nil && intent.Action == "vigil_ascend" {
			return nil
		}
		if choice != nil && slices.ContainsFunc(gachaVigil.Paths, func(path GachaVigilPath) bool { return path.aspect(*choice) != nil }) {
			return nil
		}
	}
	return ErrGachaIntent
}

func applyGachaVigil(state *GachaGameState, intent GachaActionIntent, result *GachaActionResult) error {
	if err := validateGachaVigilIntent(intent); err != nil {
		return err
	}
	own := state.Girls[*intent.GirlID]
	if !gachaCardOwned(own) || !validGachaVigil(own) {
		return ErrGachaCultivationLocked
	}
	if intent.Action == "vigil_bind" {
		if own.Vigil != nil {
			return ErrGachaCultivationLocked
		}
		own.Vigil = &GachaVigil{Path: *intent.Choice}
		result.Drops = append(result.Drops, "夜巡立誓 · "+vigilPath(*intent.Choice).Name)
		return nil
	}
	if own.Vigil == nil {
		return ErrGachaCultivationLocked
	}
	next := *own.Vigil
	path := vigilPath(next.Path)
	var evidence map[string]int
	var seals int
	var receipt string
	switch intent.Action {
	case "vigil_stance":
		if next.Overdraw == *intent.Enabled || next.Rank == 0 || *intent.Enabled && next.Strain > gachaVigil.StrainStartMax {
			return ErrGachaCultivationLocked
		}
		next.Overdraw = *intent.Enabled
		receipt = "守常：不再支付开场生命"
		if next.Overdraw {
			receipt = "借夜：下次开战将献出部分生命"
		}
	case "vigil_ground":
		if next.Strain == 0 {
			return ErrGachaCultivationLocked
		}
		evidence = map[string]int{gachaVigil.GroundEvidence: 1}
		next.Strain = max(0, next.Strain-gachaVigil.GroundDrop)
		receipt = fmt.Sprintf("锚定心神 · 失序 %d/%d", next.Strain, gachaVigil.StrainMax)
	case "vigil_ascend":
		if next.Rank >= len(gachaVigil.Ranks) || (next.Rank == 1) != (intent.Choice != nil) {
			return ErrGachaCultivationLocked
		}
		rank := gachaVigil.Ranks[next.Rank]
		if next.Marks < rank.Marks {
			return ErrGachaCultivationLocked
		}
		seals, evidence = rank.Seals, rank.Evidence
		if intent.Choice != nil {
			aspect := path.aspect(*intent.Choice)
			if aspect == nil {
				return ErrGachaCultivationLocked
			}
			next.Aspect, evidence = aspect.ID, aspect.Evidence
		}
		next.Rank++
		receipt = "仪式完成 · " + path.Name + " · " + rank.Name
	case "vigil_refocus":
		aspect := path.aspect(*intent.Choice)
		if next.Rank < 2 || aspect == nil || next.Aspect == aspect.ID {
			return ErrGachaCultivationLocked
		}
		seals, evidence, next.Aspect = gachaVigil.RefocusSeals, aspect.Evidence, aspect.ID
		receipt = "重新定相 · " + aspect.Name
	}
	if state.Mats["break_seal"] < seals {
		return ErrGachaFunds
	}
	if len(evidence) > 0 {
		stock := gachaEvidence(state)
		for key, count := range evidence {
			if stock[key] < count {
				return ErrGachaFunds
			}
		}
		for key, count := range evidence {
			stock[key] -= count
		}
	}
	if seals > 0 {
		state.Mats["break_seal"] -= seals
	}
	own.Vigil = &next
	result.Drops = append(result.Drops, receipt)
	return nil
}

type GachaVigilWitness struct {
	Path     string `json:"path"`
	Overdraw bool   `json:"overdraw,omitempty"`
}

func issueGachaVigil(state *GachaGameState, snapshot *GachaRunSnapshot) {
	if snapshot.Mode != "" {
		return
	}
	for _, id := range snapshot.Squad {
		if own := state.Girls[id]; own != nil && own.Vigil != nil {
			if snapshot.Witnesses == nil {
				snapshot.Witnesses = map[string]GachaVigilWitness{}
			}
			snapshot.Witnesses[id] = GachaVigilWitness{own.Vigil.Path, own.Vigil.Overdraw}
		}
	}
}

func settleGachaVigil(state *GachaGameState, stage *GachaStage, snapshot *GachaRunSnapshot, report *GachaBattleReport, result *GachaFinishResult) {
	if !result.Win || snapshot.Mode != "" {
		return
	}
	point := gachaVoyage.Stages[stage.ID]
	campaign := stage.RelicTier == 0 && (point == nil || point.Kind != "resources")
	for _, id := range snapshot.Squad {
		signed, ok := snapshot.Witnesses[id]
		own := state.Girls[id]
		if !ok || own == nil || own.Vigil == nil || signed.Path != own.Vigil.Path {
			continue
		}
		v := own.Vigil
		if signed.Overdraw {
			v.Strain = min(gachaVigil.StrainMax, v.Strain+gachaVigil.StrainGain)
			if v.Strain == gachaVigil.StrainMax {
				v.Overdraw = false
				result.Drops = append(result.Drops, gachaGirl(id).Name+" 触及失序上限，借夜熄灭")
			}
		} else if v.Strain > 0 {
			v.Strain--
		}
		met := v.Path == "lantern" && campaign || v.Path == "tide" && stage.RelicTier > 0 || v.Path == "ward" && report.TotalN > 0 && report.AliveN == report.TotalN
		if !met || state.Daily.VigilMarks[id] >= gachaVigil.DailyLimit || v.Marks >= gachaVigil.Ranks[len(gachaVigil.Ranks)-1].Marks {
			continue
		}
		if state.Daily.VigilMarks == nil {
			state.Daily.VigilMarks = map[string]int{}
		}
		state.Daily.VigilMarks[id]++
		v.Marks++
		result.Drops = append(result.Drops, fmt.Sprintf("%s · %s见证印 +1（%d）", gachaGirl(id).Name, vigilPath(v.Path).Name, v.Marks))
	}
}

func applyGachaVigilTeam(state *GachaGameState, units []*gachaCombatUnit) {
	aspects := map[string]map[string]bool{}
	for _, unit := range units {
		v := state.Girls[unit.GirlID].Vigil
		if v == nil || v.Rank == 0 {
			continue
		}
		if aspects[v.Path] == nil {
			aspects[v.Path] = map[string]bool{}
		}
		if v.Aspect != "" {
			aspects[v.Path][v.Aspect] = true
		}
	}
	for _, unit := range units {
		v := state.Girls[unit.GirlID].Vigil
		if v == nil || v.Rank == 0 {
			continue
		}
		value := gachaVigil.Ranks[v.Rank-1].Effect
		if v.Overdraw {
			unit.HP = max(1, unit.HP-max(1, unit.MaxHP*gachaVigil.OmenHPCost/100))
			value += gachaVigil.OmenBonus
		}
		applyGachaCultivationEffect(unit, vigilPath(v.Path).Effect, value)
		unit.vigilAspect, unit.vigilValue = v.Aspect, value
		if len(aspects) == len(gachaVigil.Paths) {
			unit.Rage = min(gachaCatalog.Formula.RageMax, unit.Rage+gachaVigil.ChorusRage)
		}
		if len(aspects[v.Path]) >= 2 {
			unit.Shield += unit.MaxHP * gachaVigil.TwinsShield / 100
		}
	}
}

func (b *gachaCombat) invokeVigil(unit *gachaCombatUnit, allies, foes []*gachaCombatUnit, trigger string) {
	if unit.vigilAspect == "" || !unit.Alive {
		return
	}
	value, id := unit.vigilValue, unit.vigilAspect
	var targets []GachaBattleTarget
	lowest := unit
	for _, ally := range allies {
		if ally.Alive && int64(ally.HP)*int64(lowest.MaxHP) < int64(lowest.HP)*int64(ally.MaxHP) {
			lowest = ally
		}
	}
	switch {
	case id == "beacon" && trigger == "cast", id == "wick" && trigger == "ready":
		if id == "wick" {
			lowest = unit
		}
		heal := min(lowest.MaxHP-lowest.HP, lowest.MaxHP*value/100)
		lowest.HP += heal
		targets = append(targets, GachaBattleTarget{UID: lowest.UID, Heal: heal})
	case id == "flare" && trigger == "strike":
		gain := min(b.formula.RageMax-unit.Rage, value)
		unit.Rage += gain
		targets = append(targets, GachaBattleTarget{UID: unit.UID, RageDelta: gain})
	case id == "anchor" && trigger == "ready":
		shield := lowest.MaxHP * value / 100
		lowest.Shield += shield
		targets = append(targets, GachaBattleTarget{UID: lowest.UID, Shield: shield})
	case id == "undertow" && trigger == "cast":
		for _, foe := range foes {
			if foe.Alive {
				loss := min(foe.Rage, value)
				foe.Rage -= loss
				targets = append(targets, GachaBattleTarget{UID: foe.UID, RageDelta: -loss})
			}
		}
	case id == "mirror" && trigger == "ready":
		unit.counter += float64(value) / 100
		targets = append(targets, GachaBattleTarget{UID: unit.UID})
	default:
		return
	}
	unit.vigilAspect = ""
	for _, path := range gachaVigil.Paths {
		if aspect := path.aspect(id); aspect != nil {
			b.record(GachaBattleEvent{Kind: "aura", Wave: b.wave, Round: b.rounds, Src: unit.UID, Skill: true, SkillName: aspect.Name, Targets: targets, DurationMS: 330})
			return
		}
	}
}
