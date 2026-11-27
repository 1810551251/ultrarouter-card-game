package model

import (
	_ "embed"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

type GachaCaseOption struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Text     string `json:"text"`
	Clue     string `json:"clue"`
	Amount   int    `json:"amount"`
	Pressure int    `json:"pressure"`
	Relief   string `json:"relief,omitempty"`
}

type GachaCaseAct struct {
	Title   string            `json:"title"`
	Text    string            `json:"text"`
	Options []GachaCaseOption `json:"options"`
}

type GachaCaseFile struct {
	Name  string         `json:"name"`
	Intro string         `json:"intro"`
	Focus string         `json:"focus"`
	Acts  []GachaCaseAct `json:"acts"`
}

type GachaCaseRules struct {
	DailyLimit    int                      `json:"daily_limit"`
	DailyBudget   int                      `json:"daily_budget"`
	PressureLimit int                      `json:"pressure_limit"`
	FileLimit     int                      `json:"file_limit"`
	PursuitLimit  int                      `json:"pursuit_limit"`
	PartialLimit  int                      `json:"partial_limit"`
	Evidence      map[string]string        `json:"evidence"`
	Endings       map[string]string        `json:"endings"`
	Oaths         map[string]string        `json:"oaths"`
	Order         []string                 `json:"order"`
	Files         map[string]GachaCaseFile `json:"files"`
}

type GachaCaseRun struct {
	ID       string         `json:"id"`
	CaseID   string         `json:"case_id"`
	Leader   string         `json:"leader,omitempty"`
	Oath     string         `json:"oath,omitempty"`
	OathUsed bool           `json:"oath_used,omitempty"`
	Trail    []string       `json:"trail"`
	Pressure int            `json:"pressure"`
	Clues    map[string]int `json:"clues"`
}

type GachaCaseReport struct {
	Date     string         `json:"date"`
	CaseID   string         `json:"case_id"`
	Ending   string         `json:"ending"`
	Pressure int            `json:"pressure"`
	Awards   map[string]int `json:"awards"`
}

type GachaCasebook struct {
	Date     string                    `json:"date"`
	Opened   []string                  `json:"opened"`
	Awarded  int                       `json:"awarded"`
	Evidence map[string]int            `json:"evidence"`
	Active   *GachaCaseRun             `json:"active,omitempty"`
	Endings  map[string]map[string]int `json:"endings"`
	Last     *GachaCaseReport          `json:"last,omitempty"`
}

//go:embed gacha_cases.json
var gachaCasesJSON []byte

var gachaCases = func() GachaCaseRules {
	var rules GachaCaseRules
	if err := common.Unmarshal(gachaCasesJSON, &rules); err != nil {
		panic(err)
	}
	if rules.DailyLimit != 2 || rules.DailyBudget != 10 || rules.PressureLimit != 7 || rules.PartialLimit < 1 || rules.FileLimit <= rules.PartialLimit || rules.PursuitLimit <= rules.FileLimit || rules.PursuitLimit > rules.DailyBudget || len(rules.Order) < 4 || len(rules.Order) != len(rules.Files) {
		panic("invalid case limits")
	}
	for _, id := range []string{"ember", "echo", "key"} {
		if rules.Evidence[id] == "" {
			panic("missing case evidence")
		}
	}
	for _, ending := range []string{"revealed", "filed", "partial", "overrun", "withdrawn"} {
		if rules.Endings[ending] == "" {
			panic("missing case ending")
		}
	}
	for index, id := range rules.Order {
		file := rules.Files[id]
		if slices.Contains(rules.Order[:index], id) || file.Name == "" || file.Intro == "" || rules.Evidence[file.Focus] == "" || len(file.Acts) != 3 {
			panic("invalid case file")
		}
		for _, act := range file.Acts {
			if act.Title == "" || act.Text == "" || len(act.Options) != 3 {
				panic("invalid case act")
			}
			for pos, opt := range act.Options {
				if opt.ID != []string{"survey", "veil", "dare"}[pos] || opt.Name == "" || opt.Text == "" || rules.Evidence[opt.Clue] == "" || opt.Amount < 1 || opt.Amount > 3 || opt.Pressure < -2 || opt.Pressure > 4 || opt.Relief != "" && rules.Evidence[opt.Relief] == "" {
					panic("invalid case option")
				}
			}
		}
	}
	return rules
}()

func gachaEvidence(state *GachaGameState) map[string]int {
	if state.Casebook == nil {
		state.Casebook = &GachaCasebook{Date: state.Daily.Date, Opened: []string{}}
	}
	if state.Casebook.Evidence == nil {
		state.Casebook.Evidence = map[string]int{}
	}
	if state.Casebook.Endings == nil {
		state.Casebook.Endings = map[string]map[string]int{}
	}
	return state.Casebook.Evidence
}

func gachaCaseEndingCount(state *GachaGameState, caseID, ending string) int {
	if state == nil || state.Casebook == nil {
		return 0
	}
	return state.Casebook.Endings[caseID][ending]
}

func validGachaCaseAmounts(values map[string]int, maximum int) bool {
	for id, count := range values {
		if gachaCases.Evidence[id] == "" || count < 0 || count > maximum {
			return false
		}
	}
	return true
}

func gachaCaseDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

func normalizeGachaCasebook(state *GachaGameState) error {
	gachaEvidence(state)
	book := state.Casebook
	if !gachaCaseDate(state.Daily.Date) || !gachaCaseDate(book.Date) || book.Date > state.Daily.Date || len(book.Opened) > gachaCases.DailyLimit || book.Awarded < 0 || book.Awarded > gachaCases.DailyBudget || !validGachaCaseAmounts(book.Evidence, gachaMaxResource) {
		return ErrGachaRunState
	}
	for index, id := range book.Opened {
		if gachaCases.Files[id].Name == "" || slices.Contains(book.Opened[:index], id) {
			return ErrGachaRunState
		}
	}
	for id, counts := range book.Endings {
		if gachaCases.Files[id].Name == "" {
			return ErrGachaRunState
		}
		for ending, count := range counts {
			if gachaCases.Endings[ending] == "" || count < 0 || count > gachaMaxResource {
				return ErrGachaRunState
			}
		}
	}
	if last := book.Last; last != nil {
		if !gachaCaseDate(last.Date) || last.Date > book.Date || gachaCaseEndingCount(state, last.CaseID, last.Ending) < 1 || last.Pressure < 0 || last.Pressure > gachaCases.PressureLimit || !validGachaCaseAmounts(last.Awards, gachaCases.PursuitLimit) {
			return ErrGachaRunState
		}
		total := 0
		for _, quantity := range last.Awards {
			total += quantity
		}
		if total > gachaCases.PursuitLimit || last.Date == book.Date && total > book.Awarded {
			return ErrGachaRunState
		}
	}
	if active := book.Active; active != nil {
		if !validateGachaUUID(active.ID) || !slices.Contains(book.Opened, active.CaseID) || active.Leader != "" && gachaGirl(active.Leader) == nil || active.Oath != "" && (active.Leader == "" || gachaCases.Oaths[active.Oath] == "") || active.Clues == nil || active.Trail == nil || len(active.Trail) > 3 {
			return ErrGachaRunState
		}
		copy := GachaCaseRun{CaseID: active.CaseID, Oath: active.Oath, Clues: map[string]int{}}
		for _, choice := range active.Trail {
			if _, ok := gachaCaseStep(&copy, choice); !ok || copy.Pressure >= gachaCases.PressureLimit {
				return ErrGachaRunState
			}
		}
		if copy.Pressure != active.Pressure || copy.OathUsed != active.OathUsed || !maps.Equal(copy.Clues, active.Clues) {
			return ErrGachaRunState
		}
	}
	if book.Date != state.Daily.Date {
		if book.Active != nil {
			if err := settleGachaCase(book, book.Active, "withdrawn", 0, nil); err != nil {
				return err
			}
		}
		book.Date, book.Opened, book.Awarded = state.Daily.Date, []string{}, 0
	}
	return nil
}

func validateGachaCaseIntent(intent GachaActionIntent) error {
	if intent.ItemID == nil || intent.Materials != nil || intent.Enabled != nil || intent.Config != nil || intent.Squad != nil || intent.OpponentID != nil || intent.QueueID != nil || intent.RoomID != nil || intent.StakeE != nil {
		return ErrGachaIntent
	}
	if intent.Action == "case_open" {
		if intent.Choice != nil || intent.Count != nil || gachaCases.Files[*intent.ItemID].Name == "" || intent.GirlID != nil && gachaGirl(*intent.GirlID) == nil {
			return ErrGachaIntent
		}
		return nil
	}
	if !validateGachaUUID(*intent.ItemID) || intent.GirlID != nil || intent.Count == nil || intent.Choice == nil {
		return ErrGachaIntent
	}
	switch intent.Action {
	case "case_step":
		if *intent.Count >= 1 && *intent.Count <= 3 && slices.Contains([]string{"survey", "veil", "dare"}, *intent.Choice) {
			return nil
		}
	case "case_close":
		if *intent.Count >= 0 && *intent.Count <= 3 && slices.Contains([]string{"file", "pursue", "withdraw"}, *intent.Choice) {
			return nil
		}
	}
	return ErrGachaIntent
}

func gachaCasePressure(run *GachaCaseRun, delta int) {
	run.Pressure = min(gachaCases.PressureLimit, max(0, run.Pressure+delta))
	if run.Oath == "ward" && !run.OathUsed && run.Pressure >= gachaCases.PressureLimit {
		run.Pressure, run.OathUsed = gachaCases.PressureLimit-2, true
	}
}

func gachaCaseStep(run *GachaCaseRun, choice string) (GachaCaseOption, bool) {
	acts := gachaCases.Files[run.CaseID].Acts
	if len(run.Trail) >= len(acts) {
		return GachaCaseOption{}, false
	}
	for _, option := range acts[len(run.Trail)].Options {
		if option.ID != choice {
			continue
		}
		amount, pressure := option.Amount, option.Pressure
		if option.Relief != "" && run.Clues[option.Relief] > 0 {
			pressure--
		}
		if !run.OathUsed && run.Oath == "lantern" && choice == "survey" {
			amount++
			run.OathUsed = true
		}
		if !run.OathUsed && run.Oath == "tide" && choice == "dare" {
			pressure--
			run.OathUsed = true
		}
		gachaCasePressure(run, pressure)
		run.Clues[option.Clue] += amount
		run.Trail = append(run.Trail, choice)
		return option, true
	}
	return GachaCaseOption{}, false
}

func settleGachaCase(book *GachaCasebook, run *GachaCaseRun, ending string, limit int, result *GachaActionResult) error {
	remaining := min(limit, gachaCases.DailyBudget-book.Awarded)
	order := []string{gachaCases.Files[run.CaseID].Focus, "ember", "echo", "key"}
	awards := map[string]int{}
	total := 0
	for remaining > 0 {
		before := remaining
		for index, id := range order {
			if remaining > 0 && !slices.Contains(order[:index], id) && run.Clues[id] > awards[id] {
				awards[id]++
				total++
				remaining--
			}
		}
		if before == remaining {
			break
		}
	}
	for id, count := range awards {
		if book.Evidence[id] > gachaMaxResource-count {
			return ErrGachaCapacity
		}
	}
	if book.Endings[run.CaseID] == nil {
		book.Endings[run.CaseID] = map[string]int{}
	}
	book.Endings[run.CaseID][ending] = min(gachaMaxResource, book.Endings[run.CaseID][ending]+1)
	for id, count := range awards {
		book.Evidence[id] += count
	}
	book.Awarded += total
	book.Last = &GachaCaseReport{Date: book.Date, CaseID: run.CaseID, Ending: ending, Pressure: run.Pressure, Awards: awards}
	book.Active = nil
	if result != nil {
		result.Drops = append(result.Drops, fmt.Sprintf("%s · %s", gachaCases.Files[run.CaseID].Name, gachaCases.Endings[ending]))
		for _, id := range slices.Sorted(maps.Keys(awards)) {
			result.Drops = append(result.Drops, fmt.Sprintf("%s ×%d", gachaCases.Evidence[id], awards[id]))
		}
	}
	return nil
}

func applyGachaCaseAction(state *GachaGameState, intent GachaActionIntent, result *GachaActionResult) error {
	if err := validateGachaCaseIntent(intent); err != nil {
		return err
	}
	if err := normalizeGachaCasebook(state); err != nil {
		return err
	}
	book := state.Casebook
	if intent.Action == "case_open" {
		id := *intent.ItemID
		if book.Active != nil || len(book.Opened) >= gachaCases.DailyLimit || slices.Contains(book.Opened, id) {
			return ErrGachaCultivationLocked
		}
		run := &GachaCaseRun{ID: strings.ToLower(intent.RequestID), CaseID: id, Clues: map[string]int{}, Trail: []string{}}
		if intent.GirlID != nil {
			if !gachaCardOwned(state.Girls[*intent.GirlID]) {
				return ErrGachaCultivationLocked
			}
			run.Leader = *intent.GirlID
			run.Oath = gachaCaseOath(state.Girls[run.Leader])
		}
		book.Active = run
		book.Opened = append(book.Opened, id)
		result.Drops = append(result.Drops, "立案 · "+gachaCases.Files[id].Name)
		return nil
	}
	active := book.Active
	if active == nil || active.ID != *intent.ItemID {
		return ErrGachaCultivationLocked
	}
	run := *active
	run.Clues = maps.Clone(active.Clues)
	run.Trail = slices.Clone(active.Trail)
	if intent.Action == "case_step" {
		if len(run.Trail)+1 != *intent.Count {
			return ErrGachaCultivationLocked
		}
		option, ok := gachaCaseStep(&run, *intent.Choice)
		if !ok {
			return ErrGachaIntent
		}
		if run.Pressure >= gachaCases.PressureLimit {
			return settleGachaCase(book, &run, "overrun", 0, result)
		}
		book.Active = &run
		result.Drops = append(result.Drops, fmt.Sprintf("%s · %s +%d · 压力 %d/%d", option.Name, gachaCases.Evidence[option.Clue], run.Clues[option.Clue]-active.Clues[option.Clue], run.Pressure, gachaCases.PressureLimit))
		return nil
	}
	if len(run.Trail) != *intent.Count {
		return ErrGachaCultivationLocked
	}
	switch *intent.Choice {
	case "withdraw":
		return settleGachaCase(book, &run, "withdrawn", 0, result)
	case "file":
		if len(run.Trail) < 1 {
			return ErrGachaCultivationLocked
		}
		return settleGachaCase(book, &run, "filed", gachaCases.FileLimit, result)
	case "pursue":
		if len(run.Trail) != 3 {
			return ErrGachaCultivationLocked
		}
		gachaCasePressure(&run, 2)
		if run.Pressure >= gachaCases.PressureLimit {
			return settleGachaCase(book, &run, "overrun", 0, result)
		}
		if len(run.Clues) >= 2 && run.Clues[gachaCases.Files[run.CaseID].Focus] >= 2 {
			total := 0
			for _, count := range run.Clues {
				total += count
			}
			if total >= 5 {
				return settleGachaCase(book, &run, "revealed", gachaCases.PursuitLimit, result)
			}
		}
		return settleGachaCase(book, &run, "partial", gachaCases.PartialLimit, result)
	}
	return ErrGachaIntent
}
