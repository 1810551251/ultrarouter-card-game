const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const {JSDOM} = require("jsdom");

function mailHarness(context) {
  const dom = new JSDOM('<body data-page="page-mail"><b data-harbor-mail-badge></b><main id="harbor-mail-body"></main></body>',{runScripts:"outside-only"});
  context.after(() => dom.window.close());
  const window = dom.window;
  window.SAVE = {mailbox:{messages:[{id:"parcel-1",title:"<礼包>",sender:"邮务署",body:"可领取的奖励",created_at:1795759200000,resources:{gold:300,stamina:30},captain_xp:40,claimed:false,read:false}],redeemed:{}}};
  window.Store = {sessionEpoch:4,ready:() => true,profile:{items:{items:{}}}};
  window.GIRLS = []; window.MATS = {};
  window.esc = value => String(value).replaceAll("&","&amp;").replaceAll("<","&lt;").replaceAll('"',"&quot;");
  window.renderTopbar = () => {};
  window.requireGame = () => true; window.closeModal = () => {}; window.nav = () => {};
  window.eval(fs.readFileSync(path.join(__dirname,"../js/mail.js"),"utf8"));
  window.renderHarborMail();
  return {window,document:window.document,message:window.SAVE.mailbox.messages[0]};
}

test("mail attachments remain pending until the server receipt, with a single ID-only request", async context => {
  const {window,document,message} = mailHarness(context);
  let finish; const calls = [];
  window.Store.mutate = (action,fields) => { calls.push(JSON.parse(JSON.stringify({action,...fields}))); return new Promise(resolve => finish=resolve); };
  assert.equal(document.querySelector("[data-harbor-mail-badge]").textContent,"1");
  assert.equal(document.querySelector(".mail-letter h3").textContent,"<礼包>");
  const pending = window.harborMailAction("mail_claim",{item_id:message.id});
  await window.harborMailAction("mail_claim",{item_id:message.id});
  assert.deepEqual(calls,[{action:"mail_claim",item_id:"parcel-1"}]);
  assert.equal(message.claimed,false); assert.equal(document.querySelector("[data-mail-claim]").disabled,true);
  message.claimed = true; finish({drops:["金币 +300"]}); await pending;
  assert.equal(document.querySelector("[data-harbor-mail-badge]").hidden,true);
  assert.match(document.querySelector(".mail-status").textContent,/金币/);
});

test("gift codes send text only, denied claims do not invent rewards, stale accounts are ignored", async context => {
  const {window,document,message} = mailHarness(context);
  let sent;
  window.Store.mutate = async (action,fields) => { sent = JSON.parse(JSON.stringify({action,...fields})); return null; };
  window.openHarborMail("redeem"); window.renderHarborMail();
  document.querySelector("#harbor-code").value = "  STARFROST2026  ";
  document.querySelector("form").requestSubmit(); await new Promise(resolve => setImmediate(resolve));
  assert.deepEqual(sent,{action:"redeem_code",item_id:"STARFROST2026"});
  assert.equal(message.claimed,false); assert.equal(window.SAVE.mailbox.messages.length,1);
  let finish;
  window.Store.mutate = () => new Promise(resolve => finish=resolve);
  const pending = window.harborMailAction("mail_all");
  window.Store.sessionEpoch++; finish({drops:["旧账号奖励"]}); await pending;
  assert.doesNotMatch(document.querySelector(".mail-status").textContent,/旧账号/);
});
