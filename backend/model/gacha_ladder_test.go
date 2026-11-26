package model_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Invoked inside the existing disposable SQLite/MySQL/PostgreSQL host matrix.
// No substitute account model or in-memory database is used by these cases.
func runGachaLadderMatrix(t *testing.T, clock time.Time) {
	t.Run("ranked_arena_21h_material_settlement", func(t *testing.T) {
		zone := time.FixedZone("Asia/Shanghai", 8*3600)
		local := clock.In(zone)
		now := time.Date(local.Year(), local.Month(), local.Day(), 20, 59, 0, 0, zone)
		tables := []any{&model.GachaLadderGate{}, &model.GachaLadderEntry{}, &model.GachaLadderMail{}, &model.GachaLadderBattle{}}
		require.NoError(t, model.DB.Migrator().DropTable(tables...))
		user := gachaUser(t)
		state := gachaState(t, user.Id, now)
		strongGachaState(state)
		seedGachaState(t, user.Id, state)
		var before model.GachaPlayer
		require.NoError(t, model.DB.First(&before, "user_id = ?", user.Id).Error)
		require.NoError(t, model.DB.AutoMigrate(tables...))
		require.NoError(t, model.InitializeGachaLadder(now))
		require.NoError(t, model.DB.AutoMigrate(tables...))
		require.NoError(t, model.InitializeGachaLadder(now))
		var after model.GachaPlayer
		require.NoError(t, model.DB.First(&after, "user_id = ?", user.Id).Error)
		assert.Equal(t, before, after, "fresh/upgrade ladder migration preserves game resources")
		var count int64
		require.NoError(t, model.DB.Model(&model.GachaLadderEntry{}).Count(&count).Error)
		assert.EqualValues(t, 100, count, "idempotent NPC initialization")
		view, err := model.GetGachaLadder(user.Id, now)
		require.NoError(t, err)
		assert.Nil(t, view.Me)
		assert.Empty(t, view.Mail)
		join := model.GachaActionIntent{RequestID: uuid.NewString(), Action: "ladder_join"}
		_, err = model.ApplyGachaAction(user.Id, join, now)
		require.NoError(t, err)
		_, err = model.ApplyGachaAction(user.Id, join, now)
		require.NoError(t, err, "replayed registration does not append a second seat")
		view, err = model.GetGachaLadder(user.Id, now)
		require.NoError(t, err)
		assert.Equal(t, 101, view.Me.Rank)
		assert.Len(t, view.Opponents, 5)
		assert.Equal(t, 5, view.Remaining)
		assert.Equal(t, 21, time.UnixMilli(view.NextSettlement).In(zone).Hour())
		assert.Equal(t, 5, time.UnixMilli(view.NextReset).In(zone).Hour())
		assert.Equal(t, view.Me.Rank, view.Me.BestRank)

		other := gachaUser(t)
		gachaState(t, other.Id, now)
		_, err = model.ApplyGachaAction(other.Id, model.GachaActionIntent{RequestID: uuid.NewString(), Action: "ladder_join"}, now)
		require.NoError(t, err)
		outsider := gachaUser(t)
		gachaState(t, outsider.Id, now)
		for _, handler := range []func(){
			func() { assert.Equal(t, 403, gachaRequest(t, user.Id, false, controller.GetGachaLadder, nil).Code) },
			func() {
				assert.Equal(t, 403, gachaRequest(t, user.Id, false, controller.GetGachaLadderReplay, nil).Code)
			},
		} {
			handler()
		}
		for _, extra := range []string{"save", "rank", "materials", "win", "report"} {
			response := gachaRequest(t, user.Id, true, controller.ApplyGachaAction, map[string]any{"request_id": uuid.NewString(), "action": "ladder_join", extra: 1})
			assert.Equal(t, 400, response.Code)
		}
		stake := 1
		assert.ErrorIs(t, model.ValidateGachaAction(model.GachaActionIntent{RequestID: uuid.NewString(), Action: "ladder_join", StakeE: &stake}), model.ErrGachaIntent)
		for _, action := range []string{"arena", "arena_join", "shop"} {
			target := -100
			assert.Error(t, model.ValidateGachaAction(model.GachaActionIntent{RequestID: uuid.NewString(), Action: action, OpponentID: &target}))
		}

		target, rank := -100, 100
		challenge := model.GachaActionIntent{RequestID: uuid.NewString(), Action: "ladder_challenge", OpponentID: &target, Count: &rank}
		out, err := model.ApplyGachaAction(user.Id, challenge, now.Add(time.Second))
		require.NoError(t, err)
		require.NotNil(t, out.Ladder)
		require.True(t, out.Ladder.Win)
		assert.Equal(t, 101, out.Ladder.OwnBefore)
		assert.Equal(t, 100, out.Ladder.OwnAfter)
		assert.Zero(t, out.Save.Daily.ArenaPlayed, "ranked challenges do not consume E matchmaking allowance")
		retry, err := model.ApplyGachaAction(user.Id, challenge, now.Add(2*time.Second))
		require.NoError(t, err)
		assert.Equal(t, out, retry)
		_, err = model.GetGachaLadderReplay(outsider.Id, out.Ladder.ID)
		assert.Error(t, err, "unrelated users cannot read private battle events")
		replay, err := model.GetGachaLadderReplay(user.Id, out.Ladder.ID)
		require.NoError(t, err)
		assert.Equal(t, out.Ladder.Report, replay.Report)
		defendedRank := 100
		defenseBattle, err := model.ApplyGachaAction(other.Id, model.GachaActionIntent{RequestID: uuid.NewString(), Action: "ladder_challenge", OpponentID: &user.Id, Count: &defendedRank}, now.Add(3*time.Second))
		require.NoError(t, err)
		require.False(t, defenseBattle.Ladder.Win, "weak attack cannot beat the registered strong defense")
		defenderReplay, err := model.GetGachaLadderReplay(user.Id, defenseBattle.Ladder.ID)
		require.NoError(t, err)
		assert.True(t, defenderReplay.Win)
		assert.Equal(t, 100, defenderReplay.OwnBefore)
		assert.Equal(t, 100, defenderReplay.OwnAfter)
		assert.Equal(t, defenseBattle.Ladder.Report, defenderReplay.Report, "both viewers replay the same attacker-left event stream")
		challenge.RequestID = uuid.NewString()
		_, err = model.ApplyGachaAction(user.Id, challenge, now.Add(2*time.Second))
		assert.ErrorIs(t, err, model.ErrGachaLadderLimit)
		_, err = model.ApplyGachaAction(user.Id, challenge, now.Add(12*time.Second))
		assert.ErrorIs(t, err, model.ErrGachaLadder, "stale rank quote must not consume an attempt")
		view, err = model.GetGachaLadder(user.Id, now.Add(12*time.Second))
		require.NoError(t, err)
		assert.Equal(t, 4, view.Remaining)

		// Ordinary squad edits must not silently replace the registered defense.
		defense := view.Me.Defense
		squad := []string{"c04"}
		_, err = model.ApplyGachaAction(user.Id, model.GachaActionIntent{RequestID: uuid.NewString(), Action: "squad", Squad: &squad}, now.Add(15*time.Second))
		require.NoError(t, err)
		view, err = model.GetGachaLadder(user.Id, now.Add(15*time.Second))
		require.NoError(t, err)
		assert.Equal(t, defense, view.Me.Defense)
		_, err = model.ApplyGachaAction(user.Id, model.GachaActionIntent{RequestID: uuid.NewString(), Action: "ladder_defense"}, now.Add(16*time.Second))
		require.NoError(t, err)
		view, err = model.GetGachaLadder(user.Id, now.Add(16*time.Second))
		require.NoError(t, err)
		assert.Len(t, view.Me.Squad, 1)

		// No worker ran at 21:00. A later challenge must pay the old rank first.
		boundary := now.Add(time.Minute)
		target, rank = -99, 99
		challenge.RequestID = uuid.NewString()
		out, err = model.ApplyGachaAction(user.Id, challenge, boundary.Add(time.Second))
		require.NoError(t, err)
		require.True(t, out.Ladder.Win)
		assert.Equal(t, 99, out.Ladder.OwnAfter)
		view, err = model.GetGachaLadder(user.Id, boundary.Add(2*time.Second))
		require.NoError(t, err)
		require.Len(t, view.Mail, 1)
		assert.Equal(t, 100, view.Mail[0].Rank, "21:00 snapshot precedes the post-21:00 rank change")
		assert.Equal(t, model.GachaLadderReward(100), view.Mail[0].Materials)
		require.NoError(t, model.SettleGachaLadder(context.Background(), boundary.Add(3*time.Second)))
		require.NoError(t, model.SettleGachaLadder(context.Background(), boundary.Add(3*time.Second)))
		day := view.Mail[0].Day
		claim := model.GachaActionIntent{RequestID: uuid.NewString(), Action: "ladder_claim", ItemID: &day}
		oldR := out.Save.Mats["mat_r"]
		claimed, err := model.ApplyGachaAction(user.Id, claim, boundary.Add(4*time.Second))
		require.NoError(t, err)
		assert.Equal(t, oldR+40, claimed.Save.Mats["mat_r"])
		again, err := model.ApplyGachaAction(user.Id, claim, boundary.Add(5*time.Second))
		require.NoError(t, err)
		assert.Equal(t, claimed, again, "claim retries do not credit twice")
		claim.RequestID = uuid.NewString()
		_, err = model.ApplyGachaAction(user.Id, claim, boundary.Add(6*time.Second))
		assert.ErrorIs(t, err, model.ErrGachaLadder)
		wallet := model.User{}
		require.NoError(t, model.DB.First(&wallet, user.Id).Error)
		assert.Equal(t, user.Quota, wallet.Quota, "ranked play and materials do not mutate E balance")

		for i := 0; i < 3; i++ {
			target, rank = -98-i, 98-i
			challenge.RequestID = uuid.NewString()
			_, err = model.ApplyGachaAction(user.Id, challenge, boundary.Add(time.Duration(20+i*12)*time.Second))
			require.NoError(t, err)
		}
		challenge.RequestID = uuid.NewString()
		_, err = model.ApplyGachaAction(user.Id, challenge, boundary.Add(time.Minute))
		assert.ErrorIs(t, err, model.ErrGachaLadderLimit)
		reset := boundary.Add(8 * time.Hour)
		view, err = model.GetGachaLadder(user.Id, reset.Add(-time.Second))
		require.NoError(t, err)
		assert.Zero(t, view.Remaining)
		view, err = model.GetGachaLadder(user.Id, reset)
		require.NoError(t, err)
		assert.Equal(t, 5, view.Remaining, "attempts reset at 05:00 Beijing, not 21:00")

		// Offline catch-up and concurrent claiming stay exactly once.
		later := boundary.Add(48 * time.Hour)
		require.NoError(t, model.SettleGachaLadder(context.Background(), later))
		view, err = model.GetGachaLadder(user.Id, later)
		require.NoError(t, err)
		require.Len(t, view.Mail, 3)
		newDay := later.In(zone).Format("2006-01-02")
		state = gachaState(t, user.Id, later)
		oldR = state.Mats["mat_r"]
		var errs [2]error
		var wg sync.WaitGroup
		for i := range errs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, errs[i] = model.ApplyGachaAction(user.Id, model.GachaActionIntent{RequestID: uuid.NewString(), Action: "ladder_claim", ItemID: &newDay}, later)
			}(i)
		}
		wg.Wait()
		assert.NotEqual(t, errs[0] == nil, errs[1] == nil)
		assert.Equal(t, oldR+40, gachaState(t, user.Id, later).Mats["mat_r"])
		previousDay := later.Add(-24 * time.Hour).In(zone).Format("2006-01-02")
		state = gachaState(t, user.Id, later)
		state.Mats["mat_r"] = 1_000_000_000
		seedGachaState(t, user.Id, state)
		_, err = model.ApplyGachaAction(user.Id, model.GachaActionIntent{RequestID: uuid.NewString(), Action: "ladder_claim", ItemID: &previousDay}, later)
		assert.ErrorIs(t, err, model.ErrGachaCapacity)
		var mail model.GachaLadderMail
		require.NoError(t, model.DB.First(&mail, "user_id = ? AND day = ?", user.Id, previousDay).Error)
		assert.False(t, mail.Claimed, "capacity failure rolls back entitlement consumption")
		require.NoError(t, model.InitializeGachaLadder(later))
		var positions []int
		require.NoError(t, model.DB.Model(&model.GachaLadderEntry{}).Order("position ASC").Pluck("position", &positions).Error)
		for i, position := range positions {
			assert.Equal(t, i+1, position, "swaps and re-initialization keep unique contiguous ranks")
		}
		require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", user.Id).Update("status", common.UserStatusDisabled).Error)
		_, err = model.GetGachaLadder(user.Id, later)
		assert.Error(t, err)
		_, err = model.GetGachaLadderReplay(user.Id, out.Ladder.ID)
		assert.Error(t, err)
	})
}
