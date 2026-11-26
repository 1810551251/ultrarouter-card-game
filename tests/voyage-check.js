const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const {JSDOM} = require("jsdom");
const read = name => fs.readFileSync(path.join(__dirname,"..",name),"utf8");
const legacy = JSON.parse(read("backend/model/gacha_catalog.json")).STAGES;

function sampleVoyage() {
  const book = {stages:{}, chapters:[], resources:["res_manifest"], first_clear:[{item:"field_notes",amount:1}], first_three:[{item:"expedition_atlas",amount:1}]};
  for (let chapter = 1; chapter <= 2; chapter++) {
    const region = {id:`c${chapter}`,name:`海域${chapter}`,landmark:`灯塔${chapter}`,routes:{},chests:{}};
    for (const [mode,count] of [["story",8],["hard",3],["nightmare",2]]) {
      const ids = [];
      for (let index=0; index<count; index++) {
        const id = mode === "story" && index === 0 ? chapter === 1 ? "s1" : "s4" : `v${chapter}${mode[0]}${index+1}`;
        ids.push(id);
        book.stages[id] = {...legacy[0], id, name:`${chapter}区${mode}${index+1}`, kind:mode, captain_min:mode === "hard" ? 8 : mode === "nightmare" ? 15 : 0, requires:index ? [ids[index-1]] : [], prize:[{item:"gear_rivet",amount:2}], sweep_prize:mode === "nightmare" ? undefined : [{item:"forge_ore",amount:2}]};
      }
      region.routes[mode] = ids;
      region.chests[mode] = [1,2,3].map(tier => ({id:`c${chapter}-${mode}-${tier}`,stars:count*tier,prize:[{item:"gold",amount:150*tier}]}));
    }
    book.chapters.push(region);
  }
  book.stages.s4.chapter_gate = "v1s8";
  book.stages.res_manifest = {...legacy[0],id:"res_manifest",name:"领航抄录所",kind:"resources",cost:6,requires:["s1"],daily_limit:3,prize:[{item:"field_notes",amount:2}], sweep_prize:[{item:"field_notes",amount:2}]};
  return book;
}

function voyageHarness(context) {
  const dom = new JSDOM(read("index.html"), {runScripts:"outside-only"});
  context.after(() => dom.window.close());
  const doc = dom.window.document, book = sampleVoyage();
  const save = {gold:700, token:20, stamina:120, staminaMax:120, squad:["c08"], girls:{}, mats:{}, daily:{date:"2026-11-27", voyage_attempts:{}}, captain:{level:7}, cleared:{}, stars:{}, voyage_claims:{}};
  dom.window.eval(["config","data","store","ui","voyage"].map(name => read(`js/${name}.js`)).join("\n;\n") + `\nconst BAT=null;function requireGame(){return Store.ready()} function renderTopbar(){} function toast(){} function closeOnly(){closeModal()} function nav(page){document.body.dataset.page=page;if(page==='page-stage')renderStage()} function startBattle(id){window.started=id} window.mapAPI={Store,openVoyage,renderStage,renderVoyage,voyageResultSummary,voyageStage,voyageCoordinates,closeOnly,setup(state){SAVE=state}};`);
  const api = dom.window.mapAPI, find = selector => doc.querySelector(selector), all = selector => [...doc.querySelectorAll(selector)];
  api.Store.ready = () => true; api.Store.profile = {voyage:book,items:JSON.parse(read("backend/model/gacha_items.json"))}; api.setup(save);
  const calls = []; api.Store.mutate = async (action, fields) => {calls.push([action,JSON.parse(JSON.stringify(fields))]);return null;};
  doc.body.dataset.page = "page-stage"; api.renderStage();
  return {dom,doc,api,save,book,calls,find,all};
}

test("voyage draws navigable desktop and portrait routes, independent modes and chapter gates", context => {
  const {api,save,find,all} = voyageHarness(context);
  assert.equal(all("[data-voy-point]").length,8); assert.equal(all("[data-voy-mode]").length,3); assert.equal(all("[data-voy-chest]").length,3);
  find('[data-voy-step="1"]').click(); assert.match(find(".voy-status").textContent,/完成前章/); assert.match(find(".voy-heading").textContent,/海域1/);
  save.voyage_passes = {s4:true}; find('[data-voy-step="1"]').click(); assert.match(find(".voy-heading").textContent,/海域2/);
  find('[data-voy-step="-1"]').click(); find('[data-voy-mode="hard"]').click(); assert.equal(all("[data-voy-point]").length,3);
  find("[data-voy-point]").click(); assert.equal(find("[data-voy-start]").disabled,true); assert.match(find(".voy-gate").textContent,/Lv\.8/);
  api.closeOnly(); find('[data-voy-mode="nightmare"]').click(); assert.equal(all("[data-voy-point]").length,2);
  assert.equal(new Set(api.voyageCoordinates(8,true).map(pair => pair[0])).size,2);
  assert.equal(new Set(api.voyageCoordinates(8,true).map(pair => pair[1])).size,4);
  assert.notDeepEqual(api.voyageCoordinates(8,false),api.voyageCoordinates(8,true));
});

test("locked resource previews name the fixed loot and direct links issue only the selected stage", context => {
  const {api,find,save,dom} = voyageHarness(context);
  api.openVoyage("resources","res_manifest");
  assert.equal(find("[data-voy-detail]").dataset.voyDetail,"res_manifest"); assert.equal(find("[data-voy-start]").disabled,true); assert.match(find(".voy-detail").textContent,/斥候札记.*×2/s);
  save.cleared.s1=true; save.daily.voyage_attempts.res_manifest=3; api.openVoyage("resources","res_manifest"); assert.match(find(".voy-gate").textContent,/零点刷新/);
  save.daily.voyage_attempts.res_manifest=0; api.openVoyage("resources","res_manifest"); find("[data-voy-start]").click(); assert.equal(dom.window.started,"res_manifest");
  assert.equal(save.stamina,120); assert.deepEqual(save.mats,{}); assert.deepEqual(save.stars,{});
  assert.equal(api.voyageStage("res_manifest").name,"领航抄录所");
});

test("chest requires stars in the selected chapter and difficulty; posts only its stable claim key", async context => {
  const {save,api,find,calls} = voyageHarness(context);
  Object.assign(save.stars,{s1:3,v1s2:3,v1s3:2,v1h1:3});
  api.renderStage(); find('[data-voy-chest="c1-story-1"]').click(); await find("[data-voy-claim]").onclick();
  assert.deepEqual(calls,[["voyage_chest",{item_id:"c1-story-1"}]]); assert.equal(save.gold,700); assert.deepEqual(save.voyage_claims,{});
  save.voyage_claims["c1-story-1"]=true; api.closeOnly(); api.renderStage(); find('[data-voy-chest="c1-story-1"]').click(); assert.equal(find("[data-voy-claim]").disabled,true);
  api.closeOnly(); find('[data-voy-mode="hard"]').click(); find('[data-voy-chest="c1-hard-2"]').click(); assert.equal(find("[data-voy-claim]").disabled,true); assert.match(find("[data-voy-claim]").textContent,/还差 3 星/);
  assert.match(api.voyageResultSummary("s1",{win:true,save}),/8\/24/);
});

test("sweeps require three stars, available attempts and stamina and submit only the dated server intent", async context => {
  const {save,api,find,calls} = voyageHarness(context);
  save.cleared.s1=true; save.stars.res_manifest=2; api.openVoyage("resources","res_manifest");
  assert.equal(find('[data-voy-sweep="1"]').disabled,true); assert.match(find(".voy-sweep").textContent,/不发站点额度/);
  save.stars.res_manifest=3; save.stamina=12; api.openVoyage("resources","res_manifest"); assert.equal(find('[data-voy-sweep="1"]').disabled,false); assert.equal(find('[data-voy-sweep="3"]').disabled,true);
  save.stamina=120; save.daily.voyage_attempts.res_manifest=1; api.openVoyage("resources","res_manifest"); assert.equal(find('[data-voy-sweep="3"]').disabled,true);
  save.daily.voyage_attempts.res_manifest=0; api.Store.activeRun={id:"in_progress"}; api.openVoyage("resources","res_manifest"); assert.equal(find('[data-voy-sweep="1"]').disabled,true);
  api.Store.activeRun=null; api.openVoyage("resources","res_manifest"); await find('[data-voy-sweep="3"]').onclick();
  assert.deepEqual(calls,[["voyage_sweep",{item_id:"res_manifest",count:3,choice:"2026-11-27"}]]); assert.equal(save.stamina,120); assert.equal(save.daily.voyage_attempts.res_manifest,0);
});
