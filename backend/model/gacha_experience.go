package model

import "fmt"

const gachaExperienceLootVersion = 7

type GachaExperienceMaterial struct {
	ID string `json:"id"`
	XP int    `json:"xp"`
}

type GachaExperienceRules struct {
	VeteranLevel  int                       `json:"veteran_level"`
	VeteranGrowth int                       `json:"veteran_growth"`
	MasterLevel   int                       `json:"master_level"`
	MasterGrowth  int                       `json:"master_growth"`
	Materials     []GachaExperienceMaterial `json:"materials"`
	Drops         map[string]map[string]int `json:"drops"`
}

type GachaFeedReceipt struct {
	LevelBefore int            `json:"level_before"`
	LevelAfter  int            `json:"level_after"`
	Experience  int            `json:"experience"`
	Spent       map[string]int `json:"spent"`
}

func gachaPeakLevel(own *GachaOwnedGirl) int {
	return max(own.Level, own.PeakLevel)
}

func gachaCardOwned(own *GachaOwnedGirl) bool {
	return own != nil && own.Star > 0 && own.Level >= 0 && own.Level <= gachaCatalog.Growth.MaxLevel &&
		(own.Level > 0 || own.Training != nil && own.Training.Grade > 0 && own.Training.Grade < len(gachaTraining.Grades) && own.PeakLevel >= gachaTraining.Grades[own.Training.Grade].Level)
}

func gachaLevelCost(level int) int {
	if level >= gachaCatalog.Growth.MaxLevel {
		return 0
	}
	rules := gachaTraining.Experience
	veteran, master := max(0, level-rules.VeteranLevel+1), max(0, level-rules.MasterLevel+1)
	return max(1, level)*gachaCatalog.Growth.ExpPerLevel + veteran*veteran*rules.VeteranGrowth + master*master*rules.MasterGrowth
}

func trainGachaLevel(own *GachaOwnedGirl, amount int) error {
	if !gachaCardOwned(own) || amount < 0 || amount > gachaMaxResource || own.Exp < 0 || own.Exp > gachaMaxResource-amount {
		return ErrGachaRunState
	}
	own.Exp += amount
	for own.Level < gachaCatalog.Growth.MaxLevel && own.Exp >= gachaLevelCost(own.Level) {
		own.Exp -= gachaLevelCost(own.Level)
		own.Level++
	}
	if own.Level == gachaCatalog.Growth.MaxLevel {
		own.Exp = 0
	}
	if own.PeakLevel > 0 {
		own.PeakLevel = gachaPeakLevel(own)
	}
	return nil
}

func validateGachaFeedIntent(intent GachaActionIntent) error {
	if intent.GirlID == nil || gachaGirl(*intent.GirlID) == nil || intent.Count != nil || intent.Squad != nil || intent.Enabled != nil || intent.ItemID != nil || intent.OpponentID != nil || intent.StakeE != nil || intent.QueueID != nil || intent.RoomID != nil || intent.Choice != nil || intent.Config != nil || len(intent.Materials) == 0 || len(intent.Materials) > len(gachaTraining.Experience.Materials) {
		return ErrGachaIntent
	}
	known := 0
	for _, material := range gachaTraining.Experience.Materials {
		if count, ok := intent.Materials[material.ID]; ok {
			if count < 1 || count > gachaMaxResource {
				return ErrGachaIntent
			}
			known++
		}
	}
	if known != len(intent.Materials) {
		return ErrGachaIntent
	}
	return nil
}

func applyGachaFeed(state *GachaGameState, intent GachaActionIntent, result *GachaActionResult) error {
	if err := validateGachaFeedIntent(intent); err != nil {
		return err
	}
	own := state.Girls[*intent.GirlID]
	if !gachaCardOwned(own) || own.Level >= gachaCatalog.Growth.MaxLevel {
		return ErrGachaCultivationLocked
	}
	for id, count := range intent.Materials {
		if state.Mats[id] < count {
			return ErrGachaFunds
		}
	}
	room := -own.Exp
	for level := own.Level; level < gachaCatalog.Growth.MaxLevel; level++ {
		room += gachaLevelCost(level)
	}
	if room <= 0 || own.Exp < 0 {
		return ErrGachaRunState
	}
	receipt := &GachaFeedReceipt{LevelBefore: own.Level, Spent: map[string]int{}}
	for _, material := range gachaTraining.Experience.Materials {
		count := min(intent.Materials[material.ID], (room+material.XP-1)/material.XP)
		if count > 0 {
			receipt.Spent[material.ID] = count
			xp := min(room, count*material.XP)
			receipt.Experience += xp
			room -= xp
		}
	}
	if err := trainGachaLevel(own, receipt.Experience); err != nil {
		return err
	}
	for id, count := range receipt.Spent {
		state.Mats[id] -= count
	}
	receipt.LevelAfter = own.Level
	result.Feed = receipt
	result.Drops = append(result.Drops, fmt.Sprintf("批量研读 · Lv.%d → Lv.%d · 实得 %d 经验", receipt.LevelBefore, receipt.LevelAfter, receipt.Experience))
	return nil
}

func awardGachaExperienceMaterials(state *GachaGameState, stage *GachaStage, result *GachaFinishResult) error {
	drops := gachaTraining.Experience.Drops[gachaMaterialStageKind(stage)]
	for id, count := range drops {
		if state.Mats[id] > gachaMaxResource-count {
			return ErrGachaCapacity
		}
	}
	for _, material := range gachaTraining.Experience.Materials {
		if count := drops[material.ID]; count > 0 {
			state.Mats[material.ID] += count
			result.Drops = append(result.Drops, fmt.Sprintf("%s ×%d", gachaItems.Items[material.ID].Name, count))
		}
	}
	return nil
}
