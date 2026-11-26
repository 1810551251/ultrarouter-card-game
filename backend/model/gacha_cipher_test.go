package model

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cipherTestGuess(board *GachaCipherState, guess string) GachaActionIntent {
	tag, index := cipherTag(board), len(board.Attempts)+1
	return GachaActionIntent{RequestID: uuid.NewString(), Action: "cipher_guess", ItemID: &tag, Count: &index, Choice: &guess}
}

func TestGachaCipherFeedbackAndPrivatePersistence(t *testing.T) {
	for _, sample := range []struct {
		answer, guess    string
		exact, misplaced int
	}{{"0011", "1100", 0, 4}, {"0112", "1011", 1, 2}, {"0125", "0000", 1, 0}, {"0011", "0011", 4, 0}, {"1122", "3333", 0, 0}} {
		assert.Equal(t, GachaCipherAttempt{Code: sample.guess, Exact: sample.exact, Misplaced: sample.misplaced}, scoreGachaCipher(sample.answer, sample.guess))
	}
	base := time.Date(2026, time.November, 26, 23, 55, 0, 0, gameZone)
	state := newGachaState(base)
	require.NoError(t, syncGachaCipher(state, base))
	raw, err := encodeGachaCipherDocument(state)
	require.NoError(t, err)
	assert.True(t, bytes.Contains(raw, []byte(`"cipher_seed"`)))
	var reload GachaGameState
	require.NoError(t, decodeGachaCipherDocument(raw, &reload))
	require.True(t, state.cipherSeed == reload.cipherSeed)
	public, err := common.Marshal(state)
	require.NoError(t, err)
	assert.False(t, bytes.Contains(public, []byte(`"cipher_seed"`)))
	assert.False(t, bytes.Contains(public, []byte(state.cipherSeed)))
	seed, err := hex.DecodeString(reload.cipherSeed)
	require.NoError(t, err)
	oldAnswer := gachaCipherAnswer(seed, reload.Cipher)
	require.NoError(t, syncGachaCipher(&reload, base.Add(5*time.Minute)))
	assert.NotEqual(t, oldAnswer, gachaCipherAnswer(seed, reload.Cipher))
	assert.Equal(t, "2026-11-27", reload.Cipher.Day)
	require.NoError(t, syncGachaCipher(&reload, base))
	assert.Equal(t, "2026-11-27", reload.Cipher.Day)
	stale := cipherTestGuess(reload.Cipher, "0123")
	assert.ErrorIs(t, applyGachaCipher(&reload, stale, &GachaActionResult{}, base), ErrGachaCultivationLocked)
	reload.Cipher.Attempts = []GachaCipherAttempt{{Code: gachaCipherAnswer(seed, reload.Cipher), Exact: 0}}
	assert.ErrorIs(t, syncGachaCipher(&reload, base), ErrGachaRunState)
}

func TestGachaCipherWinAtomicCapacityNoPracticeRewardAndExhaustion(t *testing.T) {
	base := time.Date(2026, time.November, 26, 15, 0, 0, 0, gameZone)
	state := newGachaState(base)
	require.NoError(t, syncGachaCipher(state, base))
	seed, err := hex.DecodeString(state.cipherSeed)
	require.NoError(t, err)
	answer := gachaCipherAnswer(seed, state.Cipher)
	wrong := "0000"
	if wrong == answer {
		wrong = "5555"
	}
	first := cipherTestGuess(state.Cipher, wrong)
	require.NoError(t, ValidateGachaAction(first))
	practice := func() GachaActionIntent {
		tag := cipherTag(state.Cipher)
		return GachaActionIntent{RequestID: uuid.NewString(), Action: "cipher_practice", ItemID: &tag}
	}
	assert.ErrorIs(t, applyGachaCipher(state, practice(), &GachaActionResult{}, base), ErrGachaCultivationLocked)
	require.NoError(t, applyGachaCipher(state, first, &GachaActionResult{}, base))
	assert.ErrorIs(t, applyGachaCipher(state, first, &GachaActionResult{}, base), ErrGachaCultivationLocked)
	assert.ErrorIs(t, applyGachaCipher(state, cipherTestGuess(state.Cipher, wrong), &GachaActionResult{}, base), ErrGachaCultivationLocked)
	solve := cipherTestGuess(state.Cipher, answer)
	state.Mats["forge_ore"] = gachaMaxResource
	assert.ErrorIs(t, applyGachaCipher(state, solve, &GachaActionResult{}, base), ErrGachaCapacity)
	assert.False(t, state.Cipher.Solved)
	assert.Len(t, state.Cipher.Attempts, 1)
	assert.Zero(t, state.Mats["training_ink"])
	state.Mats["forge_ore"] = 0
	require.NoError(t, applyGachaCipher(state, solve, &GachaActionResult{}, base))
	assert.Equal(t, 2, state.Mats["training_ink"])
	assert.Equal(t, 1, state.Mats["forge_ore"])
	assert.ErrorIs(t, applyGachaCipher(state, solve, &GachaActionResult{}, base), ErrGachaCultivationLocked)
	old := practice()
	require.NoError(t, applyGachaCipher(state, old, &GachaActionResult{}, base))
	assert.ErrorIs(t, applyGachaCipher(state, old, &GachaActionResult{}, base), ErrGachaCultivationLocked)
	trainingAnswer := gachaCipherAnswer(seed, state.Cipher)
	require.NotEqual(t, answer, trainingAnswer)
	require.NoError(t, applyGachaCipher(state, cipherTestGuess(state.Cipher, trainingAnswer), &GachaActionResult{}, base))
	assert.Equal(t, 2, state.Mats["training_ink"])
	assert.Equal(t, 1, state.Mats["forge_ore"])
	require.NoError(t, applyGachaCipher(state, practice(), &GachaActionResult{}, base))
	trainingAnswer = gachaCipherAnswer(seed, state.Cipher)
	for digit := 0; len(state.Cipher.Attempts) < 4; digit++ {
		guess := strings.Repeat(strconv.Itoa(digit), 4)
		if guess != trainingAnswer {
			require.NoError(t, applyGachaCipher(state, cipherTestGuess(state.Cipher, guess), &GachaActionResult{}, base))
		}
	}
	assert.False(t, state.Cipher.Solved)
	assert.ErrorIs(t, ValidateGachaAction(cipherTestGuess(state.Cipher, trainingAnswer)), ErrGachaIntent)
	require.NoError(t, applyGachaCipher(state, practice(), &GachaActionResult{}, base))
	assert.Equal(t, 3, state.Cipher.Round)
}

func TestGachaCipherRejectsForgedFieldsAndInvalidStocks(t *testing.T) {
	valid := cipherTestGuess(&GachaCipherState{Day: "2026-11-26"}, "0013")
	for _, tag := range []string{"", "2026-02-31:0", "2026-11-26:00", "2026-11-26:-1", "2026-11-26:1000000001"} {
		bad := valid
		bad.ItemID = &tag
		assert.ErrorIs(t, ValidateGachaAction(bad), ErrGachaIntent)
	}
	for _, guess := range []string{"", "123", "12345", "0560", "１２３４"} {
		bad := valid
		bad.Choice = &guess
		assert.ErrorIs(t, ValidateGachaAction(bad), ErrGachaIntent)
	}
	bad := valid
	bad.Materials = map[string]int{"forge_ore": 99}
	assert.ErrorIs(t, ValidateGachaAction(bad), ErrGachaIntent)
	bad = valid
	bad.GirlID = valid.Choice
	assert.ErrorIs(t, ValidateGachaAction(bad), ErrGachaIntent)
	bad = valid
	bad.Count = nil
	assert.ErrorIs(t, ValidateGachaAction(bad), ErrGachaIntent)
	state := &GachaGameState{Cipher: &GachaCipherState{Day: "2026-11-26", Round: 1}}
	assert.ErrorIs(t, syncGachaCipher(state, time.Now()), ErrGachaRunState)
}
