const assert = require("node:assert/strict");
const test = require("node:test");
const fs = require("node:fs");
const path = require("node:path");
const {JSDOM} = require("jsdom");
const read = file => fs.readFileSync(path.join(__dirname,"..",file),"utf8");

function setup(context) {
  const dom = new JSDOM('<div id="modal-mask"><main></main></div>',{runScripts:"outside-only"});
  context.after(() => dom.window.close());
  const rules = JSON.parse(read("backend/model/gacha_vigil.json"));
  const save = {girls:{c01:{star:2},c02:{star:1},c03:{star:2},c04:{star:2}},squad:["c01"],mats:{break_seal:10},daily:{},casebook:{evidence:{ember:2,echo:2,key:2}}};
  const view = {tab:"vigil",id:"c01",epoch:1}, calls = [];
  Object.assign(dom.window,{SAVE:save,view,Store:{sessionEpoch:1,ready:() => true,profile:{vigil:rules,cases:{evidence:{ember:"余烬拓片",echo:"回声蜡片",key:"门纹钥印"}}},mutate:async (action,fields) => { calls.push([action,JSON.parse(JSON.stringify(fields))]); return null; }},currentWorkshop:() => view,esc:String,toast:() => {},refreshCultivationViews:() => {}});
  dom.window.eval(read("js/vigil.js") + `\nfunction paintWorkshop(){document.querySelector('main').innerHTML=vigilWorkshopPanel(view.id,view.vigilChoice)}\nwindow.vt={paintWorkshop,vigilQuote,vigilAction,selectVigilPath,selectVigilAspect};paintWorkshop();`);
  return {save,view,calls,api:dom.window.vt,store:dom.window.Store,dom:dom.window.document};
}

test("night watch previews three routes without binding and charges no local state on rejected requests",async context => {
  const {api,save,calls,dom} = setup(context);
  assert.equal(dom.querySelectorAll('[data-vigil-path]').length,3);
  api.selectVigilPath("ward"); await api.vigilAction("vigil_stance",true); await api.vigilAction("vigil_refocus");
  assert.equal(calls.length,0); await api.vigilAction("vigil_bind");
  assert.deepEqual(calls,[["vigil_bind",{girl_id:"c01",choice:"ward"}]]); assert.equal(save.girls.c01.vigil,undefined);
  save.girls.c01.vigil = {path:"ward",rank:0,marks:2}; api.paintWorkshop();
  assert.equal(dom.querySelectorAll('[data-vigil-path]:disabled').length,3);
  assert.equal(dom.querySelector('[data-vigil-action="vigil_ascend"]').disabled,true);
  save.girls.c01.vigil.marks = 3; await api.vigilAction("vigil_ascend");
  assert.deepEqual(calls.at(-1),["vigil_ascend",{girl_id:"c01"}]); assert.equal(save.mats.break_seal,10);
});

test("rank two chooses an affordable aspect, rank three checks all evidence, retuning never changes oath",async context => {
  const {api,save,calls,dom} = setup(context);
  const card = save.girls.c01.vigil = {path:"lantern",rank:1,marks:21}; api.paintWorkshop();
  await api.vigilAction("vigil_ascend"); assert.equal(calls.length,0);
  assert.equal(dom.querySelectorAll('[data-vigil-aspect]').length,2); assert.ok(dom.querySelector('[data-vigil-cases]'));
  api.selectVigilAspect("anchor"); await api.vigilAction("vigil_ascend"); assert.equal(calls.length,0);
  api.selectVigilAspect("beacon"); save.casebook.evidence.ember = 1;
  await api.vigilAction("vigil_ascend"); assert.equal(calls.length,0);
  save.casebook.evidence.ember = 2; await api.vigilAction("vigil_ascend");
  assert.deepEqual(calls.at(-1),["vigil_ascend",{girl_id:"c01",choice:"beacon"}]);
  Object.assign(card,{rank:2,aspect:"beacon"}); api.selectVigilAspect("flare"); save.mats.break_seal = 5;
  await api.vigilAction("vigil_refocus"); assert.equal(calls.length,1);
  save.mats.break_seal = 6; await api.vigilAction("vigil_refocus"); assert.deepEqual(calls.at(-1),["vigil_refocus",{girl_id:"c01",choice:"flare"}]);
  save.mats.break_seal = 9; save.casebook.evidence.key = 1; await api.vigilAction("vigil_ascend"); assert.equal(calls.length,2);
  save.casebook.evidence.key = 2; await api.vigilAction("vigil_ascend"); assert.deepEqual(calls.at(-1),["vigil_ascend",{girl_id:"c01"}]);
});

test("strain restrictions and formations require eligible ranks, active teammates and distinct branches",async context => {
  const {api,save,calls,dom} = setup(context);
  const card = save.girls.c01.vigil = {path:"lantern",rank:2,aspect:"beacon",marks:10,strain:3};
  save.girls.c02.vigil = {path:"lantern",rank:2,aspect:"flare"}; save.girls.c03.vigil = {path:"ward",rank:1}; save.girls.c04.vigil = {path:"tide",rank:0}; save.squad=Object.keys(save.girls);
  api.paintWorkshop(); assert.equal(dom.querySelector('[data-vigil-formation="twins"]').dataset.active,"true");
  assert.equal(dom.querySelector('[data-vigil-formation="trine"]').dataset.active,"false");
  assert.equal(dom.querySelector('[data-vigil-stance="overdraw"]').disabled,true); await api.vigilAction("vigil_stance",true); assert.equal(calls.length,0);
  save.casebook.evidence.echo = 0; await api.vigilAction("vigil_ground"); assert.equal(calls.length,0);
  save.casebook.evidence.echo = 1; await api.vigilAction("vigil_ground"); assert.deepEqual(calls.at(-1),["vigil_ground",{girl_id:"c01"}]);
  card.strain = 2; save.girls.c04.vigil.rank = 1; api.paintWorkshop();
  assert.match(dom.querySelector('[data-vigil-mode="overdraw"]').textContent,/12%/);
  assert.equal(dom.querySelector('[data-vigil-formation="trine"]').dataset.active,"true");
  await api.vigilAction("vigil_stance",true); assert.deepEqual(calls.at(-1),["vigil_stance",{girl_id:"c01",enabled:true}]);
  assert.equal(card.overdraw,undefined); card.overdraw = true; await api.vigilAction("vigil_stance",false);
  assert.deepEqual(calls.at(-1),["vigil_stance",{girl_id:"c01",enabled:false}]);
});
