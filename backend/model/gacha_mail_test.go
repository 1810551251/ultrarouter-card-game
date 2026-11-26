package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mailTestIntent(action, id string) GachaActionIntent {
	intent := GachaActionIntent{Action: action, RequestID: uuid.NewString()}
	if id != "" {
		intent.ItemID = &id
	}
	return intent
}

func TestGachaMailGiftIsServerDefinedAndRedeemsOnce(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, gameZone)
	state := newGachaState(now)
	require.NoError(t, syncGachaCaptain(state))
	beforeGold, beforeStamina := state.Gold, state.Stamina
	intent := mailTestIntent("redeem_code", " starfrost2026 ")
	require.NoError(t, ValidateGachaAction(intent))
	receipt := &GachaActionResult{}
	require.NoError(t, applyGachaMail(state, intent, receipt, now))
	require.Len(t, state.Mailbox.Messages, 1)
	mail := state.Mailbox.Messages[0]
	assert.Equal(t, beforeGold, state.Gold)
	assert.Equal(t, beforeStamina, state.Stamina)
	assert.Zero(t, mail.CaptainXP)
	assert.Equal(t, 30, mail.Resources["stamina"])
	assert.ErrorIs(t, applyGachaMail(state, intent, receipt, now), ErrGachaRedeemCode)
	assert.ErrorIs(t, applyGachaMail(state, mailTestIntent("redeem_code", "NO-SUCH-GIFT"), receipt, now), ErrGachaRedeemCode)
	assert.ErrorIs(t, applyGachaMail(state, mailTestIntent("mail_claim", "other-user-mail"), receipt, now), ErrGachaMailUnavailable)
	require.NoError(t, applyGachaMail(state, mailTestIntent("mail_claim", mail.ID), receipt, now))
	assert.Equal(t, beforeGold+300, state.Gold)
	assert.Equal(t, beforeStamina+30, state.Stamina)
	assert.True(t, mail.Read)
	assert.True(t, mail.Claimed)
	assert.ErrorIs(t, applyGachaMail(state, mailTestIntent("mail_claim", mail.ID), receipt, now), ErrGachaMailUnavailable)
	assert.Equal(t, 1, state.Captain.Level)
	assert.Zero(t, state.Captain.XP)
	forged := 500
	intent.Count = &forged
	assert.ErrorIs(t, ValidateGachaAction(intent), ErrGachaIntent)
}

func TestGachaMailClaimAllIsAtomicAndUnreadAttachmentsCannotBeDeleted(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, gameZone)
	state := newGachaState(now)
	require.NoError(t, syncGachaCaptain(state))
	require.NoError(t, sendGachaMail(state, "first", "首封", "", map[string]int{"gold": 100}, 40, now))
	require.NoError(t, sendGachaMail(state, "second", "次封", "", map[string]int{"gold": 200, "forge_ore": 3}, 0, now))
	receipt := &GachaActionResult{}
	require.NoError(t, applyGachaMail(state, mailTestIntent("mail_read", "first"), receipt, now))
	require.NoError(t, applyGachaMail(state, mailTestIntent("mail_delete_read", ""), receipt, now))
	require.Len(t, state.Mailbox.Messages, 2)
	state.Gold = gachaMaxResource - 299
	assert.ErrorIs(t, applyGachaMail(state, mailTestIntent("mail_all", ""), receipt, now), ErrGachaCapacity)
	assert.Equal(t, gachaMaxResource-299, state.Gold)
	assert.Zero(t, state.Mats["forge_ore"])
	assert.False(t, state.Mailbox.Messages[0].Claimed)
	assert.Zero(t, state.Captain.XP)
	state.Gold = 0
	require.NoError(t, applyGachaMail(state, mailTestIntent("mail_all", ""), receipt, now))
	assert.Equal(t, 300, state.Gold)
	assert.Equal(t, 3, state.Mats["forge_ore"])
	assert.Equal(t, GachaCaptain{Level: 2, XP: 10}, *state.Captain)
	require.NoError(t, applyGachaMail(state, mailTestIntent("mail_delete_read", ""), receipt, now))
	assert.Empty(t, state.Mailbox.Messages)
}

func TestGachaMailDailyDeliveryKeepsRewardsAcrossCalendarChanges(t *testing.T) {
	now := time.Date(2026, 11, 27, 23, 59, 0, 0, gameZone)
	state := newGachaState(now)
	require.NoError(t, syncGachaCaptain(state))
	gold := state.Gold
	require.NoError(t, deliverGachaDailyMail(state, now))
	require.NoError(t, deliverGachaDailyMail(state, now))
	require.Len(t, state.Mailbox.Messages, 1)
	assert.Equal(t, gold, state.Gold)
	assert.Zero(t, state.Captain.XP)
	yesterday := state.Mailbox.Messages[0].ID
	next := now.Add(2 * time.Minute)
	regenGachaState(state, next)
	require.NoError(t, syncGachaCaptain(state))
	require.NoError(t, deliverGachaDailyMail(state, next))
	require.Len(t, state.Mailbox.Messages, 2)
	require.NoError(t, applyGachaMail(state, mailTestIntent("mail_claim", yesterday), &GachaActionResult{}, next))
	assert.Equal(t, gold+50, state.Gold)
	assert.Equal(t, GachaCaptain{Level: 2, XP: 10}, *state.Captain)
}

func TestGachaMailboxCapacityPreservesPendingRewardsAndDoesNotBlockReading(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, gameZone)
	state := newGachaState(now)
	require.NoError(t, syncGachaCaptain(state))
	for index := range 200 {
		require.NoError(t, sendGachaMail(state, fmt.Sprintf("pending-%d", index), "附件", "", map[string]int{"gold": 1}, 0, now))
	}
	assert.ErrorIs(t, sendGachaMail(state, "overflow", "未发放", "", map[string]int{"gold": 1}, 0, now), ErrGachaMailboxFull)
	require.NoError(t, deliverGachaDailyMail(state, now))
	assert.True(t, state.Daily.CaptainLog.Claimed["rollcall"])
	require.Len(t, state.Mailbox.Messages, 201)
	next := now.AddDate(0, 0, 1)
	regenGachaState(state, next)
	require.NoError(t, syncGachaCaptain(state))
	require.NoError(t, deliverGachaDailyMail(state, next))
	require.NoError(t, sendGachaMail(state, "battle:completed", "副本结算", "", map[string]int{"forge_ore": 7}, 0, next))
	require.Len(t, state.Mailbox.Messages, 201)
	summary := state.Mailbox.Messages[200]
	assert.Equal(t, 100, summary.Resources["gold"])
	assert.Equal(t, 80, summary.CaptainXP)
	assert.Equal(t, 7, summary.Resources["forge_ore"])
	require.NoError(t, applyGachaMail(state, mailTestIntent("mail_all", ""), &GachaActionResult{}, now))
	require.NoError(t, deliverGachaDailyMail(state, next))
	assert.True(t, state.Daily.CaptainLog.Claimed["rollcall"])
}

func TestGachaMailGrowthContractDoesNotRouteTheDispatchContract(t *testing.T) {
	assert.Equal(t, "每日契约奖励", gachaMailActionTitle("daily_contract_claim"))
	assert.Empty(t, gachaMailActionTitle("contract_claim"))
}

func TestGachaMailRewardsCanBeDeliveredAtCapacityAndKeepActualStaminaCosts(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, gameZone)
	state := newGachaState(now)
	state.Progression = &GachaProgressionState{Permits: 17, Marks: 9}
	state.Gold, state.Mats["training_ink"] = gachaMaxResource, gachaMaxResource
	state.Girls["c08"].Shards = 12
	before := gachaMailableBalance(state)
	buffer := beginGachaMailRewards(state)
	drops, err := grantGachaVoyagePrize(state, []GachaVoyagePrize{{"gold", 25}, {"training_ink", 2}})
	require.NoError(t, err)
	require.NoError(t, spendGachaStamina(state, 6, now, &drops))
	require.NoError(t, mailGachaRewards(state, buffer, "action:reward-at-capacity", "副本奖励", now, &drops))
	before["stamina"] -= 6
	assert.Equal(t, before, gachaMailableBalance(state))
	assert.Equal(t, 6, state.Captain.XP)
	require.Len(t, state.Mailbox.Messages, 1)
	message := state.Mailbox.Messages[0]
	assert.Equal(t, map[string]int{"gold": 25, "training_ink": 2}, message.Resources)
	assert.ErrorIs(t, applyGachaMail(state, mailTestIntent("mail_claim", message.ID), &GachaActionResult{}, now), ErrGachaCapacity)
	assert.False(t, message.Claimed)
	assert.Equal(t, before, gachaMailableBalance(state))
	buffer = beginGachaMailRewards(state)
	require.NoError(t, mailGachaRewards(state, buffer, "battle:no-rewards", "未获得物资", now, &drops))
	assert.Equal(t, before, gachaMailableBalance(state))
	require.Len(t, state.Mailbox.Messages, 1)
}
