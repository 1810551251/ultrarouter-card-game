package model_test

import (
	"slices"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testGachaMailDatabase(t *testing.T, _ time.Time) {
	const maxResource = 1_000_000_000
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, time.FixedZone("Beijing", 8*3600))
	command := func(action, item string) model.GachaActionIntent {
		intent := model.GachaActionIntent{Action: action, RequestID: uuid.NewString()}
		if item != "" {
			intent.ItemID = &item
		}
		return intent
	}
	t.Run("gift_ownership_replay_and_deletion", func(t *testing.T) {
		owner, other := gachaUser(t), gachaUser(t)
		before, otherBefore := gachaState(t, owner.Id, now), gachaState(t, other.Id, now)
		redeem := command("redeem_code", " starfrost2026 ")
		sent, err := model.ApplyGachaAction(owner.Id, redeem, now)
		require.NoError(t, err)
		gift := requireGachaTestMail(t, sent.Save, "gift:welcome-2026")
		assert.Equal(t, map[string]int{"gold": 300, "stamina": 30, "field_notes": 2}, gift.Resources)
		assert.Zero(t, gift.CaptainXP)
		assert.False(t, gift.Claimed)
		assert.Equal(t, before.Gold, sent.Save.Gold)
		assert.Equal(t, before.Stamina, sent.Save.Stamina)
		assert.Equal(t, before.Mats, sent.Save.Mats)
		assert.Equal(t, before.Captain, sent.Save.Captain)
		again, err := model.ApplyGachaAction(owner.Id, redeem, now)
		require.NoError(t, err)
		assert.Equal(t, sent, again)
		_, err = model.ApplyGachaAction(owner.Id, command("redeem_code", "STARFROST2026"), now)
		require.ErrorIs(t, err, model.ErrGachaRedeemCode)
		claim := command("mail_claim", gift.ID)
		_, err = model.ApplyGachaAction(other.Id, claim, now)
		require.ErrorIs(t, err, model.ErrGachaMailUnavailable)
		untouched := gachaState(t, other.Id, now)
		assert.Equal(t, otherBefore.Mailbox, untouched.Mailbox)
		assert.Equal(t, otherBefore.Gold, untouched.Gold)
		assert.False(t, requireGachaTestMail(t, gachaState(t, owner.Id, now), gift.ID).Claimed)
		paid, err := model.ApplyGachaAction(owner.Id, claim, now)
		require.NoError(t, err)
		assert.True(t, requireGachaTestMail(t, paid.Save, gift.ID).Claimed)
		assert.Equal(t, before.Gold+300, paid.Save.Gold)
		assert.Equal(t, before.Stamina+30, paid.Save.Stamina)
		assert.Equal(t, before.Mats["field_notes"]+2, paid.Save.Mats["field_notes"])
		assert.Equal(t, before.Captain, paid.Save.Captain)
		again, err = model.ApplyGachaAction(owner.Id, claim, now)
		require.NoError(t, err)
		assert.Equal(t, paid, again)
		_, err = model.ApplyGachaAction(owner.Id, command("mail_claim", gift.ID), now)
		require.ErrorIs(t, err, model.ErrGachaMailUnavailable)
		cleaned, err := model.ApplyGachaAction(owner.Id, command("mail_delete_read", ""), now)
		require.NoError(t, err)
		assert.False(t, slices.ContainsFunc(cleaned.Save.Mailbox.Messages, func(message *model.GachaMailMessage) bool { return message.ID == gift.ID }))
		assert.True(t, cleaned.Save.Mailbox.Redeemed["welcome-2026"])
		_, err = model.ApplyGachaAction(owner.Id, command("redeem_code", "STARFROST2026"), now)
		require.ErrorIs(t, err, model.ErrGachaRedeemCode)
		again, err = model.ApplyGachaAction(owner.Id, redeem, now)
		require.NoError(t, err)
		assert.Equal(t, sent, again)
		current := gachaState(t, owner.Id, now)
		assert.Equal(t, cleaned.Save.Mailbox, current.Mailbox)
		assert.Equal(t, paid.Save.Gold, current.Gold)
		assert.Equal(t, paid.Save.Stamina, current.Stamina)
		assert.Equal(t, paid.Save.Mats, current.Mats)
		independent, err := model.ApplyGachaAction(other.Id, redeem, now)
		require.NoError(t, err)
		ownGift := requireGachaTestMail(t, independent.Save, gift.ID)
		assert.Equal(t, otherBefore.Gold+300, claimGachaTestMail(t, other.Id, ownGift, now).Gold)
	})
	t.Run("claim_all_capacity_rollback_and_retry", func(t *testing.T) {
		owner := gachaUser(t)
		before := gachaState(t, owner.Id, now)
		rollcall := requireGachaTestMail(t, before, "daily:2026-11-27:rollcall")
		require.Positive(t, rollcall.CaptainXP)
		for _, code := range []string{"STARFROST2026", "VOYAGEGIFT"} {
			_, err := model.ApplyGachaAction(owner.Id, command("redeem_code", code), now)
			require.NoError(t, err)
		}
		state := gachaState(t, owner.Id, now)
		require.Len(t, state.Mailbox.Messages, 3)
		gold := 400 + rollcall.Resources["gold"]
		state.Gold = 1_000_000_000 - gold + 1
		seedGachaState(t, owner.Id, state)
		all := command("mail_all", "")
		_, err := model.ApplyGachaAction(owner.Id, all, now)
		require.ErrorIs(t, err, model.ErrGachaCapacity)
		unpaid := gachaState(t, owner.Id, now)
		assert.Equal(t, state.Gold, unpaid.Gold)
		assert.Equal(t, state.Stamina, unpaid.Stamina)
		assert.Equal(t, state.Mats, unpaid.Mats)
		assert.Equal(t, state.Captain, unpaid.Captain)
		assert.Equal(t, state.Mailbox, unpaid.Mailbox)
		unpaid.Gold--
		seedGachaState(t, owner.Id, unpaid)
		paid, err := model.ApplyGachaAction(owner.Id, all, now)
		require.NoError(t, err)
		assert.Equal(t, 1_000_000_000, paid.Save.Gold)
		assert.Equal(t, before.Stamina+30, paid.Save.Stamina)
		for id, amount := range map[string]int{"field_notes": 2, "training_ink": 2, "gear_rivet": 5} {
			assert.Equal(t, before.Mats[id]+amount, paid.Save.Mats[id], id)
		}
		assert.Equal(t, &model.GachaCaptain{Level: 2, XP: 10}, paid.Save.Captain)
		for _, message := range paid.Save.Mailbox.Messages {
			assert.True(t, message.Claimed)
			assert.True(t, message.Read)
		}
		again, err := model.ApplyGachaAction(owner.Id, all, now)
		require.NoError(t, err)
		assert.Equal(t, paid, again)
		_, err = model.ApplyGachaAction(owner.Id, command("mail_all", ""), now)
		require.ErrorIs(t, err, model.ErrGachaMailUnavailable)
	})
	t.Run("star_chest_full_gold_and_material_wait_for_room", func(t *testing.T) {
		owner := gachaUser(t)
		state := gachaState(t, owner.Id, now)
		state.Gold, state.Mats["training_manual"] = maxResource, maxResource
		state.Stars["s1"], state.Stars["s5"], state.Stars["s2"] = 3, 3, 2
		seedGachaState(t, owner.Id, state)
		const chestID = "c1-story-1"
		open := command("voyage_chest", chestID)
		sent, err := model.ApplyGachaAction(owner.Id, open, now)
		require.NoError(t, err)
		assert.True(t, sent.Save.VoyageClaims[chestID])
		assert.Equal(t, maxResource, sent.Save.Gold)
		assert.Equal(t, maxResource, sent.Save.Mats["training_manual"])
		letter := requireGachaTestMail(t, sent.Save, "action:"+open.RequestID)
		assert.Equal(t, map[string]int{"gold": 150, "training_manual": 1}, letter.Resources)
		again, err := model.ApplyGachaAction(owner.Id, open, now)
		require.NoError(t, err)
		assert.Equal(t, sent, again)
		_, err = model.ApplyGachaAction(owner.Id, command("voyage_chest", chestID), now)
		require.ErrorIs(t, err, model.ErrGachaVoyageChest)
		claim := command("mail_claim", letter.ID)
		_, err = model.ApplyGachaAction(owner.Id, claim, now)
		require.ErrorIs(t, err, model.ErrGachaCapacity)
		unpaid := gachaState(t, owner.Id, now)
		assert.False(t, requireGachaTestMail(t, unpaid, letter.ID).Claimed)
		unpaid.Gold -= 150
		seedGachaState(t, owner.Id, unpaid)
		_, err = model.ApplyGachaAction(owner.Id, claim, now)
		require.ErrorIs(t, err, model.ErrGachaCapacity)
		unpaid = gachaState(t, owner.Id, now)
		assert.False(t, requireGachaTestMail(t, unpaid, letter.ID).Claimed)
		assert.Equal(t, maxResource-150, unpaid.Gold)
		assert.Equal(t, maxResource, unpaid.Mats["training_manual"])
		unpaid.Mats["training_manual"]--
		seedGachaState(t, owner.Id, unpaid)
		paid, err := model.ApplyGachaAction(owner.Id, claim, now)
		require.NoError(t, err)
		assert.True(t, requireGachaTestMail(t, paid.Save, letter.ID).Claimed)
		assert.Equal(t, maxResource, paid.Save.Gold)
		assert.Equal(t, maxResource, paid.Save.Mats["training_manual"])
		again, err = model.ApplyGachaAction(owner.Id, claim, now)
		require.NoError(t, err)
		assert.Equal(t, paid, again)
	})
	t.Run("full_store_allows_sweep_and_charges_stamina_once", func(t *testing.T) {
		const stage = "res_manifest"
		owner := gachaUser(t)
		state := gachaState(t, owner.Id, now)
		state.Cleared["s1"], state.Cleared[stage], state.Stars[stage] = true, true, 3
		state.Captain.Level, state.Mats["field_notes"] = 5, maxResource
		seedGachaState(t, owner.Id, state)
		sweep := sweepCommand(stage, "2026-11-27", 3)
		sent, err := model.ApplyGachaAction(owner.Id, sweep, now)
		require.NoError(t, err)
		letter := requireGachaTestMail(t, sent.Save, "action:"+sweep.RequestID)
		assert.Equal(t, map[string]int{"field_notes": 6, "training_manual": 3}, letter.Resources)
		assert.Equal(t, state.Stamina-18, sent.Save.Stamina)
		assert.Equal(t, state.Captain.XP+18, sent.Save.Captain.XP)
		assert.Equal(t, 3, sent.Save.Daily.VoyageAttempts[stage])
		assert.Equal(t, maxResource, sent.Save.Mats["field_notes"])
		assert.Equal(t, state.Mats["training_manual"], sent.Save.Mats["training_manual"])
		again, err := model.ApplyGachaAction(owner.Id, sweep, now)
		require.NoError(t, err)
		assert.Equal(t, sent, again)
		_, err = model.ApplyGachaAction(owner.Id, command("mail_claim", letter.ID), now)
		require.ErrorIs(t, err, model.ErrGachaCapacity)
		unpaid := gachaState(t, owner.Id, now)
		assert.Equal(t, sent.Save.Stamina, unpaid.Stamina)
		assert.Equal(t, sent.Save.Captain, unpaid.Captain)
		assert.Equal(t, sent.Save.Daily.VoyageAttempts, unpaid.Daily.VoyageAttempts)
		assert.False(t, requireGachaTestMail(t, unpaid, letter.ID).Claimed)
		var battles int64
		require.NoError(t, model.DB.Model(&model.GachaRun{}).Where("user_id = ?", owner.Id).Count(&battles).Error)
		assert.Zero(t, battles)
	})
	t.Run("winning_resource_battle_mails_full_store_and_keeps_receipt", func(t *testing.T) {
		const stage = "res_manifest"
		owner := gachaUser(t)
		state := gachaState(t, owner.Id, now)
		strongGachaState(state)
		state.Cleared["s1"], state.Mats["field_notes"] = true, maxResource
		seedGachaState(t, owner.Id, state)
		run, err := model.StartGachaRun(owner.Id, stage, uuid.NewString(), now)
		require.NoError(t, err)
		require.True(t, run.Result.Win)
		assert.Equal(t, state.Stamina-6, run.Save.Stamina)
		assert.Equal(t, state.Captain.XP+6, run.Save.Captain.XP)
		finishAt := time.UnixMilli(run.ReadyAt + 1)
		out, err := model.FinishGachaRun(owner.Id, run.ID, false, finishAt)
		require.NoError(t, err)
		require.True(t, out.Win)
		assert.True(t, out.Save.Cleared[stage])
		assert.Equal(t, run.Save.Stamina, out.Save.Stamina)
		assert.Equal(t, run.Save.Captain, out.Save.Captain)
		assert.Equal(t, maxResource, out.Save.Mats["field_notes"])
		assert.Equal(t, state.Mats["training_manual"], out.Save.Mats["training_manual"])
		assert.Zero(t, out.Quota)
		assert.Equal(t, owner.Quota, out.WalletQuota)
		letter := requireGachaTestMail(t, out.Save, "battle:"+run.ID)
		assert.Equal(t, map[string]int{"field_notes": 2, "training_manual": 1}, letter.Resources)
		_, err = model.ApplyGachaAction(owner.Id, command("mail_claim", letter.ID), finishAt)
		require.ErrorIs(t, err, model.ErrGachaCapacity)
		again, err := model.FinishGachaRun(owner.Id, run.ID, false, finishAt)
		require.NoError(t, err)
		assert.Equal(t, out, again)
		unpaid := gachaState(t, owner.Id, finishAt)
		assert.Equal(t, out.Save.Mats, unpaid.Mats)
		assert.Equal(t, out.Save.Mailbox, unpaid.Mailbox)
		assert.False(t, requireGachaTestMail(t, unpaid, letter.ID).Claimed)
	})
	t.Run("receiving_earned_shards_unlocks_only_once", func(t *testing.T) {
		owner := gachaUser(t)
		state := gachaState(t, owner.Id, now)
		cardID := "c23"
		require.Nil(t, state.Girls[cardID])
		state.Girls[cardID] = &model.GachaOwnedGirl{Shards: 49, Gems: make([]string, 3)}
		week := &state.Progression.Weeks[0]
		week.Completed, week.Activities, week.Rules.ShardsPerWeek = 5, []string{"training", "wish"}, 2
		seedGachaState(t, owner.Id, state)
		weekly := command("weekly_claim", week.Week)
		weekly.GirlID = &cardID
		sent, err := model.ApplyGachaAction(owner.Id, weekly, now)
		require.NoError(t, err)
		letter := requireGachaTestMail(t, sent.Save, "action:"+weekly.RequestID)
		assert.Equal(t, 2, letter.Resources["shard:"+cardID])
		locked := sent.Save.Girls[cardID]
		assert.Zero(t, locked.Star)
		assert.Zero(t, locked.Level)
		assert.Equal(t, 49, locked.Shards)
		claim := command("mail_claim", letter.ID)
		paid, err := model.ApplyGachaAction(owner.Id, claim, now)
		require.NoError(t, err)
		unlocked := paid.Save.Girls[cardID]
		assert.Equal(t, 2, unlocked.Star)
		assert.Equal(t, 1, unlocked.Level)
		assert.Equal(t, 1, unlocked.Shards)
		assert.True(t, requireGachaTestMail(t, paid.Save, letter.ID).Claimed)
		again, err := model.ApplyGachaAction(owner.Id, claim, now)
		require.NoError(t, err)
		assert.Equal(t, paid, again)
		_, err = model.ApplyGachaAction(owner.Id, command("mail_claim", letter.ID), now)
		require.ErrorIs(t, err, model.ErrGachaMailUnavailable)
		assert.Equal(t, unlocked, gachaState(t, owner.Id, now).Girls[cardID])
	})
}
