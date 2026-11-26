const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const { JSDOM } = require("jsdom");
const root = path.join(__dirname, "..");
const events = JSON.parse(fs.readFileSync(path.join(root,"backend/model/gacha_festival.json")));

function festivalHarness() {
  const dom = new JSDOM('<body data-page="page-festival"><div id="festival-body"></div></body>', {runScripts:"outside-only"});
  const window = dom.window;
  window.Store = {profile:{festival:{events,week:"2026-W48",album:{3:{forge_ore:10}}}},sessionEpoch:3, ready:() => true};
  window.SAVE = {daily:{date:"2026-11-23"},festival:{week:"2026-W48",day:0,stamps:{},tries:{},claimed:{}}};
  window.esc = value => String(value).replaceAll("<", "&lt;");
  window.renderTopbar = () => {};
  window.eval(fs.readFileSync(path.join(root,"js/festival.js"), "utf8"));
  window.renderFestival();
  return dom;
}

test("seven ports preview all fourteen fixtures but only today's events accept a proposal", () => {
  const dom = festivalHarness(), doc = dom.window.document;
  try {
    assert.equal(doc.querySelectorAll("[data-day]").length, 7);
    assert.equal(doc.querySelectorAll("form[data-fest-id]").length, 2);
    const names = [...doc.querySelectorAll(".fest-event h3")].map(node => node.textContent);
    for (let index = 1; index < 7; index++) {
      doc.querySelector(`[data-day="${index}"]`).click();
      assert.equal(doc.querySelectorAll(".fest-event form").length, 0);
      names.push(...[...doc.querySelectorAll(".fest-event h3")].map(node => node.textContent));
    }
    assert.equal(new Set(names).size, 14); assert.equal(doc.querySelector("[data-album]").disabled, true);
  } finally { dom.window.close(); }
});

test("port forms submit the dated proposal only and display the server award", async () => {
  const dom = festivalHarness(), window = dom.window, doc = window.document;
  try {
    let sent;
    window.Store.mutate = async (action, fields) => {
      sent = JSON.parse(JSON.stringify({action,...fields}));
      window.SAVE.festival.stamps.relief = true;
      return {save:window.SAVE,drops:["斥候札记 ×2"]};
    };
    const form = doc.querySelector('[data-fest-id="relief"]');
    assert.ok(form.querySelector("button").disabled);
    form.querySelector('input[value="0"]').click(); form.querySelector('input[value="1"]').click();
    assert.equal(form.querySelector("button").disabled, false);
    assert.match(form.querySelector("output").textContent, /重 7 · 值 15/);
    form.requestSubmit(); await new Promise(resolve => setImmediate(resolve));
    assert.deepEqual(sent,{action:"festival_play",item_id:"2026-11-23:relief",count:1,choice:"01"});
    assert.match(doc.querySelector(".fest-notice").textContent,/斥候札记/);
    assert.equal(doc.querySelector('[data-fest-id="relief"]'), null);
    window.Store.profile.festival.week = "2026-W49"; window.renderFestival();
    assert.equal(doc.querySelector("form"),null); assert.match(doc.querySelector("#festival-body").textContent,/同步/);
  } finally { dom.window.close(); }
});
