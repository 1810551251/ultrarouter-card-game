const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const {JSDOM} = require("jsdom");

function setup(context) {
  const browser = new JSDOM('<body data-page="page-captain"><span id="res-captain"></span><div id="captain-summary"></div><main id="captain-body"></main></body>',{runScripts:"outside-only"});
  context.after(() => browser.window.close());
  const log = {claimed:{},progress:{rollcall:1,voyage:2},supplies:{}};
  const state = {captain:{level:3,xp:8},daily:{date:"2026-11-27",captain_log:log}};
  const rules = {xp_base:20,xp_step:10,tasks:[{id:"rollcall",name:"点名",goal:"抵港",target:1,xp:40,gold:50},{id:"voyage",name:"巡航",goal:"赢三次",target:3,xp:100,gold:160}],supplies:[{id:"morning",name:"便当",start:9,end:12,stamina:30},{id:"noon",name:"炊烟",start:12,end:21,stamina:30},{id:"night",name:"热汤",start:21,end:24,stamina:30}]};
  let now = Date.parse("2026-11-27T01:00:00Z");
  const calls = [], notices = [], pages = [];
  const store = {sessionEpoch:1,profile:{captain_rules:rules},ready:() => true,serverNow:() => now,mutate:async(action,fields) => { calls.push([action,JSON.parse(JSON.stringify(fields))]); return null; }};
  Object.assign(browser.window,{SAVE:state,Store:store,GROWTH:{maxLevel:150},esc:String,toast:text => notices.push(text),nav:page => pages.push(page)});
  browser.window.eval(fs.readFileSync(path.join(__dirname,"../js/captain.js"),"utf8")+`\nfunction renderTopbar(){renderCaptainSummary()}\nwindow.journalTest={renderCaptain,renderCaptainSummary,captainJournalClick};renderCaptain();renderCaptainSummary();`);
  const api = browser.window.journalTest, document = browser.window.document;
  return {state,store,api,log,calls,notices,pages,document,clock:time => { now=Date.parse(time); api.renderCaptain(); api.renderCaptainSummary(); },click:selector => api.captainJournalClick({target:document.querySelector(selector)})};
}

test("captain journal reads confirmed progress and submits only dated task and meal IDs", async context => {
  const {document,log,calls,click,api,pages} = setup(context);
  assert.equal(document.querySelector("#res-captain").textContent,"Lv.3");
  assert.match(document.querySelector("#captain-summary").textContent,/1 项任务可领奖.*便当可领/);
  assert.equal(document.querySelector(".captain-xp progress").value,8);
  await click('[data-captain-visit="voyage"]');
  assert.deepEqual(pages,["page-stage"]);
  await click('[data-captain-task="rollcall"]');
  await click('[data-captain-supply="morning"]');
  await click('[data-captain-supply="night"]');
  assert.deepEqual(calls,[["captain_task",{item_id:"2026-11-27:rollcall"}],["captain_supply",{item_id:"2026-11-27:morning"}]]);
  assert.deepEqual(log.claimed,{});
  assert.deepEqual(log.supplies,{});
  log.claimed.rollcall = true; log.supplies.morning = true; api.renderCaptain();
  assert.equal(document.querySelector('[data-captain-task="rollcall"]').disabled,true);
  assert.match(document.querySelector('[data-captain-supply="morning"]').textContent,/已领取/);
});

test("meal windows and stale journal actions follow Beijing server time, including midnight", async context => {
  const {clock,document,state,store,notices,calls,click} = setup(context);
  for (const [time,open] of [["00:59:59",null],["01:00:00","morning"],["04:00:00","noon"],["13:00:00","night"],["16:00:00",null]]) {
    clock(`2026-11-27T${time}Z`);
    assert.deepEqual([...document.querySelectorAll('[data-captain-supply]:not(:disabled)')].map(button => button.dataset.captainSupply),open ? [open] : []);
  }
  assert.match(document.querySelector("#captain-summary").textContent,/新航日/);
  clock("2026-11-27T13:00:00Z");
  state.daily.date = "2026-11-28";
  await click('[data-captain-task="rollcall"]');
  assert.deepEqual(calls,[]);
  state.daily.date = "2026-11-27"; clock("2026-11-27T13:00:00Z");
  let resolve;
  store.mutate = () => new Promise(done => {resolve = done;});
  const pending = click('[data-captain-supply="night"]');
  store.sessionEpoch++; resolve({drops:["旧账号提示"]}); await pending;
  assert.deepEqual(notices,[]);
});
