const assert = require("node:assert/strict");
const test = require("node:test");
const fs = require("node:fs");
const path = require("node:path");
const {JSDOM} = require("jsdom");

function setup(context) {
  const dom = new JSDOM('<div id="modal-mask"><main></main></div>',{runScripts:"outside-only"});
  context.after(() => dom.window.close());
  const board = {day:"2026-11-26",round:0,solved:false,attempts:[]}, view = {id:"c01",tab:"dispatch",dispatchRoom:"cipher",epoch:1}, calls = [], notices = [];
  Object.assign(dom.window,{SAVE:{cipher:board},view,Store:{sessionEpoch:1,ready:() => true,mutate:async(action,fields) => {calls.push([action,JSON.parse(JSON.stringify(fields))]); return null;}},currentWorkshop:() => view,esc:String,toast:value => notices.push(value)});
  dom.window.eval(fs.readFileSync(path.join(__dirname,"../js/cipher.js"),"utf8")+`\nfunction paintWorkshop(){document.querySelector('main').innerHTML=cipherWorkshopPanel(view)}\nwindow.ct={cipherPick,cipherAction,paintWorkshop};paintWorkshop();`);
  return {board,view,calls,notices,store:dom.window.Store,api:dom.window.ct,dom:dom.window.document};
}

test("cipher choices are local drafts only; the board paints only confirmed feedback", async context => {
  const {api,board,dom,calls} = setup(context);
  assert.equal(dom.querySelectorAll('[data-cipher-symbol]').length,6); assert.equal(dom.querySelectorAll('[data-cipher-slot]').length,4);
  api.cipherPick("symbol",6); api.cipherPick("symbol",0); await api.cipherAction("cipher_guess"); assert.equal(calls.length,0);
  for (const symbol of [0,2,4]) api.cipherPick("symbol",symbol);
  await api.cipherAction("cipher_guess");
  assert.deepEqual(calls,[["cipher_guess",{item_id:"2026-11-26:0",count:1,choice:"0024"}]]);
  assert.deepEqual(board.attempts,[]); assert.equal(dom.querySelectorAll('[data-cipher-attempt]').length,0);
  board.attempts.push({code:"0024",exact:1,misplaced:2}); api.paintWorkshop();
  assert.match(dom.querySelector('[data-cipher-attempt="1"]').textContent,/同位 1 · 异位 2/);
  for (const symbol of [0,0,2,4]) api.cipherPick("symbol",symbol);
  assert.equal(dom.querySelector('[data-cipher-action="cipher_guess"]').disabled,true);
  api.cipherPick("slot",0); api.cipherPick("symbol",5); await api.cipherAction("cipher_guess");
  assert.deepEqual(calls.at(-1),["cipher_guess",{item_id:"2026-11-26:0",count:2,choice:"5024"}]);
});

test("only a finished board offers free practice, and drafts/late receipts do not cross boards or sessions",async context => {
  const {api,board,dom,calls,store,view,notices} = setup(context);
  await api.cipherAction("cipher_practice"); assert.equal(calls.length,0); board.solved = true; api.paintWorkshop();
  await api.cipherAction("cipher_practice"); assert.deepEqual(calls,[["cipher_practice",{item_id:"2026-11-26:0"}]]);
  board.round = 1; board.solved = false; api.paintWorkshop(); assert.match(dom.querySelector('[data-cipher-panel]').textContent,/余 4 猜/);
  for (const symbol of [0,1,2,3]) api.cipherPick("symbol",symbol);
  let finish; store.mutate = () => new Promise(done => { finish = done; });
  const pending = api.cipherAction("cipher_guess"); assert.ok(finish);
  view.epoch = ++store.sessionEpoch; board.day = "2026-11-27"; board.round = 0; api.paintWorkshop();
  finish({drops:["stale cipher"]}); await pending; assert.deepEqual(notices,[]);
  assert.equal(dom.querySelector('[data-cipher-action="cipher_guess"]').disabled,true);
});
