package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func festivalSolutions(event GachaFestivalEvent) (wins, misses []string) {
	width := map[string]int{"cargo": 2, "sound": 1, "relay": 3, "tide": 1}[event.Kind]
	for value := 0; value < []int{0, 10, 100, 1000}[width]; value++ {
		choice := fmt.Sprintf("%0*d", width, value)
		if valid, win := festivalGuess(event, choice); valid {
			if win {
				wins = append(wins, choice)
			} else {
				misses = append(misses, choice)
			}
		}
	}
	return wins, misses
}

func TestGachaFestivalEveryEditionPlayable(t *testing.T) {
	seen := map[string]bool{}
	for _, event := range festivalEvents {
		require.False(t, seen[event.ID])
		seen[event.ID] = true
		for id, quantity := range event.Reward {
			require.NotEmpty(t, gachaItems.Items[id].Name)
			require.Positive(t, quantity)
		}
	}
	monday := time.Date(2026, 1, 5, 0, 0, 0, 0, gameZone)
	for week := 0; week < 12; week++ {
		rules := festivalEdition(monday.AddDate(0, 0, week*7))
		require.Len(t, rules.Events, 14)
		for _, event := range rules.Events {
			wins, misses := festivalSolutions(event)
			require.NotEmpty(t, wins, event.ID)
			require.NotEmpty(t, misses, event.ID)
			if event.Kind != "cargo" {
				require.Len(t, wins, 1, event.ID)
			}
		}
	}
}

func TestGachaFestivalReceiptsAndCalendar(t *testing.T) {
	now := time.Date(2026, 11, 23, 0, 0, 0, 0, gameZone)
	state := newGachaState(now)
	result := &GachaActionResult{}
	event := festivalEdition(now).Events[0]
	wins, misses := festivalSolutions(event)
	key, attempt := gachaDay(now)+":"+event.ID, 1
	intent := GachaActionIntent{Action: "festival_play", RequestID: uuid.NewString(), ItemID: &key, Count: &attempt, Choice: &misses[0]}
	require.NoError(t, ValidateGachaAction(intent))
	require.NoError(t, applyGachaFestival(state, intent, result, now))
	require.Empty(t, result.Drops)
	require.Empty(t, state.Festival.Stamps)
	require.ErrorIs(t, applyGachaFestival(state, intent, result, now), ErrGachaMode)
	attempt++
	intent.Choice = &wins[0]
	state.Mats["field_notes"] = gachaMaxResource
	require.ErrorIs(t, applyGachaFestival(state, intent, result, now), ErrGachaCapacity)
	require.Equal(t, 1, state.Festival.Tries[event.ID])
	require.Empty(t, state.Festival.Stamps)
	state.Mats["field_notes"] = 0
	require.NoError(t, applyGachaFestival(state, intent, result, now))
	require.Equal(t, 2, state.Mats["field_notes"])
	require.ErrorIs(t, applyGachaFestival(state, intent, result, now), ErrGachaMode)
	require.ErrorIs(t, applyGachaFestival(state, intent, result, now.AddDate(0, 0, 1)), ErrGachaMode)
	week, threshold := gachaWeekKey(now), 3
	album := GachaActionIntent{Action: "festival_album", RequestID: uuid.NewString(), ItemID: &week, Count: &threshold}
	require.ErrorIs(t, applyGachaFestival(state, album, result, now), ErrGachaMode)
	state.Festival.Stamps["whisper"] = true
	state.Festival.Stamps["wards"] = true
	state.Festival.Tries["whisper"], state.Festival.Tries["wards"] = 1, 1
	require.NoError(t, applyGachaFestival(state, album, result, now))
	require.True(t, state.Festival.Claimed[3])
	require.ErrorIs(t, applyGachaFestival(state, album, result, now), ErrGachaMode)
	require.ErrorIs(t, applyGachaFestival(state, album, result, now.AddDate(0, 0, 7)), ErrGachaMode)
	require.Empty(t, state.Festival.Stamps)
	require.Empty(t, state.Festival.Claimed)
}
