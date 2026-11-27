package model

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

type GachaCipherAttempt struct {
	Code      string `json:"code"`
	Exact     int    `json:"exact"`
	Misplaced int    `json:"misplaced"`
}

type GachaCipherState struct {
	Day      string               `json:"day"`
	Round    int                  `json:"round"`
	Attempts []GachaCipherAttempt `json:"attempts"`
	Solved   bool                 `json:"solved"`
}

type gachaCipherDocument struct {
	*GachaGameState
	CipherSeed string `json:"cipher_seed,omitempty"`
}

func encodeGachaCipherDocument(state *GachaGameState) ([]byte, error) {
	return common.Marshal(gachaCipherDocument{state, state.cipherSeed})
}

func decodeGachaCipherDocument(document []byte, state *GachaGameState) error {
	record := gachaCipherDocument{GachaGameState: state}
	if err := common.Unmarshal(document, &record); err != nil {
		return err
	}
	state.cipherSeed = record.CipherSeed
	return nil
}

func cipherLimit(round int) int {
	if round == 0 {
		return 5
	}
	return 4
}

func validCipherGuess(code string) bool {
	if len(code) != 4 {
		return false
	}
	for index := range len(code) {
		if code[index] < '0' || code[index] > '5' {
			return false
		}
	}
	return true
}

func cipherTag(board *GachaCipherState) string { return board.Day + ":" + strconv.Itoa(board.Round) }

func gachaCipherAnswer(seed []byte, board *GachaCipherState) string {
	digest := hmac.New(sha256.New, seed)
	digest.Write([]byte("cipher.v1/" + cipherTag(board)))
	bytes := digest.Sum(nil)
	date, _ := time.Parse("2006-01-02", board.Day)
	parity := int((date.Unix()/86400 + int64(board.Round) + int64(seed[0])) & 1)
	return string([]byte{'0' + 2*(bytes[0]%3) + byte(parity), '0' + bytes[1]%6, '0' + bytes[2]%6, '0' + bytes[3]%6})
}

func scoreGachaCipher(answer, guess string) GachaCipherAttempt {
	feedback := GachaCipherAttempt{Code: guess}
	var remaining, wanted [6]int
	for index := range len(answer) {
		if answer[index] == guess[index] {
			feedback.Exact++
			continue
		}
		remaining[answer[index]-'0']++
		wanted[guess[index]-'0']++
	}
	for symbol := range remaining {
		feedback.Misplaced += min(remaining[symbol], wanted[symbol])
	}
	return feedback
}

func syncGachaCipher(state *GachaGameState, now time.Time) error {
	today, board := gachaDay(now), state.Cipher
	if state.cipherSeed == "" {
		if board != nil && (board.Round != 0 || len(board.Attempts) != 0 || board.Solved) {
			return ErrGachaRunState
		}
		var bytes [32]byte
		if _, err := rand.Read(bytes[:]); err != nil {
			return err
		}
		state.cipherSeed = hex.EncodeToString(bytes[:])
	}
	seed, err := hex.DecodeString(state.cipherSeed)
	if err != nil || len(seed) != 32 {
		return ErrGachaRunState
	}
	if board != nil {
		if _, err := time.Parse("2006-01-02", board.Day); err != nil || board.Round < 0 || board.Round > gachaMaxResource || len(board.Attempts) > cipherLimit(board.Round) {
			return ErrGachaRunState
		}
		answer := gachaCipherAnswer(seed, board)
		solved := false
		seen := map[string]bool{}
		for _, attempt := range board.Attempts {
			if solved || seen[attempt.Code] || !validCipherGuess(attempt.Code) || scoreGachaCipher(answer, attempt.Code) != attempt {
				return ErrGachaRunState
			}
			solved, seen[attempt.Code] = attempt.Exact == 4, true
		}
		if solved != board.Solved {
			return ErrGachaRunState
		}
	}
	if board == nil || board.Day < today {
		state.Cipher = &GachaCipherState{Day: today, Attempts: []GachaCipherAttempt{}}
	}
	return nil
}

func validateGachaCipherIntent(intent GachaActionIntent) error {
	if intent.ItemID == nil || len(*intent.ItemID) < 12 || len(*intent.ItemID) > 22 || intent.GirlID != nil || intent.Squad != nil || intent.Enabled != nil || intent.Config != nil || intent.Materials != nil || intent.OpponentID != nil || intent.QueueID != nil || intent.RoomID != nil || intent.StakeE != nil {
		return ErrGachaIntent
	}
	day, seq, ok := strings.Cut(*intent.ItemID, ":")
	if !ok {
		return ErrGachaIntent
	}
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return ErrGachaIntent
	}
	round, err := strconv.Atoi(seq)
	if err != nil || round < 0 || round > gachaMaxResource || strconv.Itoa(round) != seq {
		return ErrGachaIntent
	}
	if intent.Action == "cipher_practice" && intent.Choice == nil && intent.Count == nil {
		return nil
	}
	if intent.Action == "cipher_guess" && intent.Choice != nil && validCipherGuess(*intent.Choice) && intent.Count != nil && *intent.Count >= 1 && *intent.Count <= cipherLimit(round) {
		return nil
	}
	return ErrGachaIntent
}

func applyGachaCipher(state *GachaGameState, intent GachaActionIntent, result *GachaActionResult, now time.Time) error {
	if err := validateGachaCipherIntent(intent); err != nil {
		return err
	}
	board := state.Cipher
	if board == nil || board.Day != gachaDay(now) || *intent.ItemID != cipherTag(board) {
		return ErrGachaCultivationLocked
	}
	ended := board.Solved || len(board.Attempts) >= cipherLimit(board.Round)
	if intent.Action == "cipher_practice" {
		if !ended || board.Round >= gachaMaxResource {
			return ErrGachaCultivationLocked
		}
		state.Cipher = &GachaCipherState{Day: board.Day, Round: board.Round + 1, Attempts: []GachaCipherAttempt{}}
		result.Drops = append(result.Drops, "新练习密本已开启 · 无材料奖励")
		return nil
	}
	guess := *intent.Choice
	if ended || *intent.Count != len(board.Attempts)+1 || slices.ContainsFunc(board.Attempts, func(attempt GachaCipherAttempt) bool { return attempt.Code == guess }) {
		return ErrGachaCultivationLocked
	}
	seed, err := hex.DecodeString(state.cipherSeed)
	if err != nil || len(seed) != 32 {
		return ErrGachaRunState
	}
	feedback := scoreGachaCipher(gachaCipherAnswer(seed, board), guess)
	if feedback.Exact == 4 && board.Round == 0 {
		if state.Mats == nil || state.Mats["training_ink"] > gachaMaxResource-2 || state.Mats["forge_ore"] >= gachaMaxResource {
			return ErrGachaCapacity
		}
		state.Mats["training_ink"] += 2
		state.Mats["forge_ore"]++
		result.Drops = append(result.Drops, "译出日刊密本 · 研训墨 ×2 · 星锻矿 ×1")
	} else {
		result.Drops = append(result.Drops, fmt.Sprintf("同位 %d · 异位 %d", feedback.Exact, feedback.Misplaced))
	}
	board.Attempts = append(board.Attempts, feedback)
	board.Solved = feedback.Exact == 4
	return nil
}
