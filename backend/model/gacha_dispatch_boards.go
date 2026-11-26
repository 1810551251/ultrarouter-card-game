package model

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

func dispatchLaneOK(track []int, lane int) bool {
	if lane < 0 || lane > 2 || len(track) >= 4 {
		return false
	}
	if len(track) == 0 {
		return true
	}
	last := track[len(track)-1]
	return lane >= last-1 && lane <= last+1
}
func dispatchRouteStep(route *GachaFogRoute, lane int) bool {
	if route.Status != "open" || !dispatchLaneOK(route.Track, lane) {
		return false
	}
	tile := gachaDispatch.Tiles[route.Tiles[3*len(route.Track)+lane]]
	if route.Light < tile.Fuel {
		return false
	}
	route.Light = min(7, route.Light-tile.Fuel+tile.Light)
	route.Cargo += tile.Cargo
	route.Track = append(route.Track, lane)
	if len(route.Track) == 4 {
		route.Status = "arrived"
	} else if route.Light == 0 {
		route.Status = "lost"
	}
	return true
}
func validDispatchRoute(route GachaFogRoute) bool {
	if len(route.Tiles) != 12 || len(route.Track) > 4 {
		return false
	}
	for _, id := range route.Tiles {
		if _, ok := gachaDispatch.Tiles[id]; !ok {
			return false
		}
	}
	check := GachaFogRoute{Tiles: route.Tiles, Light: 5, Status: "open"}
	for _, lane := range route.Track {
		if !dispatchRouteStep(&check, lane) {
			return false
		}
	}
	return check.Light == route.Light && check.Cargo == route.Cargo && (check.Status == route.Status || route.Status == "banked" && check.Status == "open" && len(check.Track) > 0)
}
func applyGachaRoute(state *GachaGameState, intent GachaActionIntent, result *GachaActionResult) error {
	route := &state.Daily.Dispatch.Route
	next := *route
	next.Track = slices.Clone(route.Track)
	if next.Status != "open" {
		return ErrGachaCultivationLocked
	}
	if intent.Action == "route_step" {
		if len(next.Track)+1 != *intent.Count || !dispatchRouteStep(&next, slices.Index([]string{"left", "middle", "right"}, *intent.Choice)) {
			return ErrGachaCultivationLocked
		}
	} else {
		if len(next.Track) == 0 || len(next.Track) != *intent.Count {
			return ErrGachaCultivationLocked
		}
		next.Status = "banked"
	}
	if next.Status == "banked" || next.Status == "arrived" {
		reward := map[string]int{}
		if next.Cargo > 0 {
			reward["forge_ore"] = next.Cargo
		}
		if next.Status == "arrived" {
			reward["route_badge"] = 3
		}
		if err := grantGachaDispatch(state, reward, result); err != nil {
			return err
		}
	}
	*route = next
	result.Drops = append(result.Drops, map[string]string{"open": "抵达驿点，可继续或带货返程", "lost": "灯油耗尽，未入库的货物留在了雾中", "banked": "沿原路撤回，货物已交库", "arrived": "穿过四层迷雾，驿路全程贯通"}[next.Status])
	return nil
}

func dispatchMosaicMarks(week string, page int) []int {
	if page >= 3 {
		return nil
	}
	rng := dispatchRNG(dispatchSeed("mosaic", week, fmt.Sprint(page)))
	pieces := [3][3]int{{0, 1, 3}, {2, 5, 8}, {4, 6, 7}}
	turn, mirror := rng.IntN(4), rng.IntN(2)
	marks := make([]int, 0, 3)
	for _, piece := range pieces {
		pos := piece[rng.IntN(3)]
		row, col := pos/3, pos%3
		if mirror == 1 {
			col = 2 - col
		}
		for range turn {
			row, col = col, 2-row
		}
		marks = append(marks, row*3+col)
	}
	sort.Ints(marks)
	return marks
}
func validDispatchMosaic(marks []int, layout string) bool {
	if len(marks) != 3 || len(layout) != 11 || layout[3] != '-' || layout[7] != '-' {
		return false
	}
	cells := strings.ReplaceAll(layout, "-", "")
	if len(cells) != 9 {
		return false
	}
	lines := 0
	for _, ink := range []byte{'a', 'b', 'c'} {
		var group []int
		cores := 0
		for index := range len(cells) {
			if cells[index] == ink {
				group = append(group, index)
				if slices.Contains(marks, index) {
					cores++
				}
			}
		}
		if len(group) != 3 || cores != 1 {
			return false
		}
		links := 0
		for index, pos := range group {
			for _, other := range group[index+1:] {
				if other-pos == 3 || other-pos == 1 && other/3 == pos/3 {
					links++
				}
			}
		}
		if links != 2 {
			return false
		}
		if group[0]/3 == group[2]/3 || group[0]%3 == group[1]%3 && group[1]%3 == group[2]%3 {
			lines++
		}
	}
	return lines == 1
}
func applyGachaMosaic(state *GachaGameState, intent GachaActionIntent, result *GachaActionResult) error {
	board := &state.Dispatch.Mosaic
	if *intent.ItemID != board.Week || *intent.Count != board.Cleared+1 || !validDispatchMosaic(board.Marks, *intent.Choice) {
		return ErrGachaCultivationLocked
	}
	if err := grantGachaDispatch(state, gachaDispatch.MosaicReward, result); err != nil {
		return err
	}
	board.Cleared++
	board.Marks = dispatchMosaicMarks(board.Week, board.Cleared)
	result.Drops = append(result.Drops, fmt.Sprintf("拼合星图 · 本周 %d/3 幅", board.Cleared))
	return nil
}

func applyMirrorTactic(units []*gachaCombatUnit, tactic string) {
	if len(units) != 2 {
		return
	}
	switch tactic {
	case "ward":
		units[0].Shield += units[0].MaxHP / 4
	case "spark":
		units[0].Rage = min(gachaCatalog.Formula.RageMax, units[0].Rage+65)
	case "rush":
		for _, unit := range units {
			unit.Atk += unit.Atk * 15 / 100
			unit.HP = max(1, unit.HP-unit.MaxHP/8)
		}
	}
}
func simulateDispatchMirror(scene GachaMirrorScene, day string, squad []string, tactic string) (*GachaMirrorLog, error) {
	borrowed := &GachaGameState{Squad: slices.Clone(squad), Girls: map[string]*GachaOwnedGirl{}, mirrorTactic: tactic}
	for _, id := range squad {
		if !slices.Contains(scene.Pool, id) {
			return nil, ErrGachaTeam
		}
		borrowed.Girls[id] = &GachaOwnedGirl{Star: 3, Level: 32, Gems: newGachaGemSlots()}
	}
	stage := GachaStage{EnemyLv: scene.Level, Waves: 2, EnemyCount: scene.Enemies, StarRound: 7, EnemyElem: scene.Enemy, Mechanic: scene.Mechanic}
	report, err := simulateGachaCombat(borrowed, stage, dispatchRNG(dispatchSeed("mirror-battle", day)))
	if err != nil {
		return nil, err
	}
	log := &GachaMirrorLog{Squad: slices.Clone(squad), Choice: tactic, Win: report.Win, Alive: report.AliveN, Rounds: report.Rounds, Notes: []string{}}
	for _, event := range report.Events {
		if event.Kind != "action" || !event.Skill || event.Skipped || !strings.HasPrefix(event.Src, "a") {
			continue
		}
		log.Casts++
		if len(log.Notes) < 5 {
			log.Notes = append(log.Notes, fmt.Sprintf("第 %d 波 · 第 %d 回合 · %s", event.Wave+1, event.Round, event.SkillName))
		}
	}
	return log, nil
}
func mirrorObjective(scene GachaMirrorScene, log *GachaMirrorLog) bool {
	return log.Win && (scene.Goal == "survive" && log.Alive == 2 || scene.Goal == "rush" && log.Rounds <= 7 || scene.Goal == "cast" && log.Casts >= 2)
}
func applyGachaMirror(state *GachaGameState, intent GachaActionIntent, result *GachaActionResult) error {
	desk := &state.Daily.Dispatch.Mirror
	if desk.Won || desk.Attempts+1 != *intent.Count {
		return ErrGachaCultivationLocked
	}
	scene := gachaDispatch.Mirrors[desk.Scene]
	log, err := simulateDispatchMirror(scene, state.Daily.Date, *intent.Squad, *intent.Choice)
	if err != nil {
		return err
	}
	won := mirrorObjective(scene, log)
	if won {
		if err := grantGachaDispatch(state, gachaDispatch.MirrorReward, result); err != nil {
			return err
		}
	}
	desk.Attempts++
	desk.Won, desk.Last = won, log
	label := "本局未达战术目标，可依据战记调整"
	if won {
		label = "镜厅试阵达标，记录已封存"
	}
	result.Drops = append(result.Drops, label)
	return nil
}
