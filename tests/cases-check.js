const test = require("node:test");
const assert = require("node:assert/strict");
const vm = require("node:vm");
const fs = require("node:fs");
const path = require("node:path");
const read = file=>fs.readFileSync(path.join(__dirname,"..",file),"utf8");
const rules=JSON.parse(read("backend/model/gacha_cases.json"));
const plain=value=>JSON.parse(JSON.stringify(value));

function fixture() {
  const book={date:"2026-11-26",opened:[],awarded:0,evidence:{key:2},endings:{}}, requests=[], girl={id:"c01",name:"领队"}, own={star:1,vigil:{path:"lantern",rank:1}}, view={id:"c01",tab:"cases",epoch:1}, select={checked:true};
  const ctx=vm.createContext({Store:{profile:{cases:rules},sessionEpoch:1,ready:()=>true,mutate:async(action,fields)=>{requests.push([action,plain(fields)]);return null;}},SAVE:{casebook:book},esc:String,GIRLS:[girl],currentWorkshop:()=>view,document:{querySelector:()=>select}});
  const api=vm.runInContext(read("js/cases.js")+"\n;({casesWorkshopPanel,caseWorkshopAction})",ctx);
  return {book,requests,select,view,api,html:()=>api.casesWorkshopPanel(girl,own)};
}

test("four distinct dossiers have three original decision scenes each and an attainable proof objective",()=>{
  assert.equal(rules.daily_limit,2);assert.equal(rules.daily_budget,10);assert.equal(new Set(rules.order).size,4);assert.equal(Object.keys(rules.endings).length,5);
  const unique=new Set();for(const id of rules.order){const file=rules.files[id];assert.ok(rules.evidence[file.focus]);assert.equal(file.acts.length,3);for(const act of file.acts){assert.ok(!unique.has(act.text));unique.add(act.text);assert.deepEqual(act.options.map(option=>option.id),["survey","veil","dare"]);assert.ok(act.options[0].amount>act.options[1].amount);assert.ok(act.options[2].pressure>act.options[0].pressure);}}
  const {html}=fixture();const markup=html();for(const id of rules.order)assert.ok(markup.includes(`data-case-open="${id}"`));assert.match(markup,/灯使/);assert.match(markup,/data-case-evidence="echo"/);
});

test("rendered investigation choices bind to the frozen case, stage, leader and server receipt",async()=>{
  const {book,requests,select,api,html}=fixture();const before=plain(book);
  await api.caseWorkshopAction("open","silent_bell");assert.deepEqual(requests.pop(),["case_open",{item_id:"silent_bell",girl_id:"c01"}]);
  select.checked=false;await api.caseWorkshopAction("open","paper_rain");assert.deepEqual(requests.pop(),["case_open",{item_id:"paper_rain"}]);assert.deepEqual(book,before);
  const id="bd096ed7-db13-4c3b-bf1d-6c1b8120e42e";book.opened=["silent_bell"];book.active={id,case_id:"silent_bell",leader:"c01",oath:"lantern",oath_used:true,trail:["survey"],clues:{echo:3},pressure:1};
  assert.match(html(),/data-case-oath="lantern"/);assert.match(html(),/已互证，压力 −1/);assert.match(html(),/已使用/);assert.match(html(),/data-case-step="dare"/);
  await api.caseWorkshopAction("close","pursue");assert.equal(requests.length,0);
  await api.caseWorkshopAction("step","veil");assert.deepEqual(requests.pop(),["case_step",{count:2,item_id:id,choice:"veil"}]);await api.caseWorkshopAction("close","file");assert.deepEqual(requests.pop(),["case_close",{count:1,item_id:id,choice:"file"}]);
  book.active.trail.push("veil","survey");await api.caseWorkshopAction("close","pursue");assert.deepEqual(requests.pop(),["case_close",{count:3,item_id:id,choice:"pursue"}]);
  delete book.active;book.last={case_id:"silent_bell",ending:"revealed",awards:{echo:3,key:1}};book.endings.silent_bell={revealed:1};assert.match(html(),/data-case-ending="revealed"/);assert.match(html(),/回声蜡片 ×3/);
  book.opened.push("glass_clock");await api.caseWorkshopAction("open","paper_rain");assert.equal(requests.length,0);
});
