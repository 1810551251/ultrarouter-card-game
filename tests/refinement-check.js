const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");

const read = name => fs.readFileSync(path.join(__dirname,"..",name),"utf8");
const profile = Object.fromEntries([["equipment","equipment"],["cultivation","cultivation"],["items","items"]].map(([key,file]) => [key,JSON.parse(read(`backend/model/gacha_${file}.json`))]));

function fixture() {
  const card = {star:12,level:60,weapon_level:10,sockets:2,gems:["ruby","ruby@1"],chips:["edge_1@5"],equipment:{head:{level:30,star:12}}};
  const state = {girls:{c01:card},squad:["c01"],mats:{gear_rivet:230,route_badge:59,gear_plate:27,gear_thread:5,forge_ore:4},gem_inventory:{ruby:4,"ruby@2":1},chip_inventory:{edge_1:40}};
  const sandbox = vm.createContext({Store:{profile,ready:()=>true,sessionEpoch:1},state,esc:String});
  const api = vm.runInContext(["data","atelier","workshop"].map(name=>read(`js/${name}.js`)).join("\n;\n") + `\n;SAVE=state;({atelierGearGrade,atelierGearArt,atelierGearQuote,atelierGearStat,artifactIdentity,refinementItem,equippedRefinement,workshopSlotInfo,workshopSlotControls,workshopRefinementInfo,workshopItems,workshopSlotAction,setView(view){workshopView=view}})`,sandbox);
  return {state,card,api,store:sandbox.Store};
}

test("bound gear retains per-piece grades and crown while quoting both overlimit materials",()=>{
  const {api,state,card}=fixture();
  for (const [level,grade,glow] of [[0,"white",0],[1,"white",1],[2,"white",2],[3,"green",0],[6,"blue",0],[10,"purple",0],[15,"gold",0],[22,"orange",0],[30,"red",2]]) {
    assert.equal(api.atelierGearGrade(level).quality,grade);assert.equal(api.atelierGearGrade(level).glow,glow);
  }
  const old=api.atelierGearQuote("c01","weapon");assert.equal(old.level,10);assert.equal(old.grade.quality,"purple");assert.equal(api.atelierGearStat(old,10,5),32);
  let quote=api.atelierGearQuote("c01","head");assert.equal(quote.starCost,230);assert.equal(quote.markCost,60);assert.equal(quote.canStar,false);assert.match(api.atelierGearArt(quote),/data-gear-prestige="crown"/);
  state.mats.route_badge++;assert.equal(api.atelierGearQuote("c01","head").canStar,true);
  card.equipment.head.star=14;quote=api.atelierGearQuote("c01","head");assert.equal(quote.starCost,450);assert.equal(quote.markCost,160);assert.match(api.atelierGearArt(quote),/data-gear-prestige="beyond"/);
  card.equipment.head.star=15;quote=api.atelierGearQuote("c01","head");assert.equal(quote.canStar,false);assert.match(api.atelierGearArt(quote),/data-gear-prestige="zenith"/);
  assert.ok(profile.equipment.level_costs.slice(-5).reduce((sum,value)=>sum+value,0)>6*profile.equipment.level_costs.slice(0,10).reduce((sum,value)=>sum+value,0));
});

test("only equipped entries add flat stats and inventory keeps each refined level",()=>{
  const {api,card,state} = fixture();
  assert.equal(api.equippedRefinement(card).atk,8+11+15);
  state.gem_inventory["ruby@12"]=100;state.chip_inventory["edge_1@12"]=100;assert.equal(api.equippedRefinement(card).atk,34);
  assert.deepEqual(Array.from(api.workshopItems("gems",1,true,card.gems)).filter(entry=>entry.base==="ruby").map(entry=>[entry.id,entry.count]),[["ruby",4],["ruby@1",0],["ruby@2",1],["ruby@12",100]]);
  assert.equal(api.workshopItems("gems",undefined,false).length,5);assert.equal(api.workshopItems("chips",undefined,false).length,18);
  for(const [tab,keys] of [["gems",Object.keys(profile.equipment.refinement.gems)],["chips",Object.keys(profile.equipment.refinement.chips).map(family=>`${family}_1`)]]) for(const key of keys) {
    for(let level=0;level<12;level++) assert.ok(api.refinementItem(tab,`${key}@${level+1}`).value>api.refinementItem(tab,level ? `${key}@${level}` : key).value);
  }
  for (const id of ["ruby@0","ruby@01","ruby@13","edge_1@+1","ruby@2@1"]) assert.equal(api.artifactIdentity(id),null);
});

test("workshop refines the displayed equipped key, never its preview or inventory",async()=>{
  const {api,state,store,card}=fixture(), view={id:"c01",tab:"chips",slot:1,item:"edge_1",epoch:1};api.setView(view);
  const ready=()=>api.workshopSlotControls(view,api.workshopSlotInfo(view)).find(entry=>entry.action==="chip_refine").ready;
  assert.equal(ready(),true);assert.match(api.workshopRefinementInfo(view,api.workshopSlotInfo(view)),/破阵脉冲/);
  const requests=[];store.mutate=async(action,fields)=>{requests.push([action,{...fields}]);return null};const before=structuredClone(state);
  await api.workshopSlotAction("chip_refine");assert.deepEqual(requests,[["chip_refine",{girl_id:"c01",count:1,item_id:"edge_1@5"}]]);assert.deepEqual(state,before);
  card.chip_locks=1;assert.equal(ready(),false);await api.workshopSlotAction("chip_refine");assert.equal(requests.length,1);
  card.chip_locks=0;state.mats.gear_plate=26;assert.equal(ready(),false);card.chips[0]="edge_1@12";assert.equal(ready(),false);
});
