const assert = require("node:assert/strict");
const test = require("node:test");
const fs = require("node:fs");
const path = require("node:path");
const { JSDOM } = require("jsdom");

function setup(context, guild = null) {
  const dom = new JSDOM('<body data-page="page-fleet"><div id="fleet-body"></div></body>', { runScripts: "outside-only" });
  context.after(() => dom.window.close());
  const calls = [], reads = [], notices = [];
  const store = {
    ready: () => true, sessionEpoch: 1, profile: { user_id: 14, guild },
    refresh: async () => store.profile,
    mutate: async (action, fields) => { calls.push([action, JSON.parse(JSON.stringify(fields))]); return null; },
    api: async (url, method) => {
      assert.equal(method, "GET");
      const query = new URL(url, "https://example.invalid").searchParams;
      const board = query.get("board"), after = Number(query.get("after"));
      reads.push([board, after]);
      const entries = board === "directory" ? [{ name: after ? "后续" : "<img src=x>", members: 3, code: "00000002", value: 4 }] : [{ name: "藏家", mine: true, value: 7 }];
      return { board, entries, catalog_size: 28, member_limit: 10, my_rank: 78, my_value: 7, next: board === "directory" && !after ? 12 : 0 };
    },
  };
  const window = dom.window;
  Object.assign(window, { Store: store, SITE: { ep: { fleet: "/api/user/gacha/fleet" } }, ELEM: {}, SAVE: { stamina: 90, squad: ["c01"] }, requireGame: () => true, renderTopbar() {}, toast: notice => notices.push(notice) });
  const ui = fs.readFileSync(path.join(__dirname, "../js/ui.js"), "utf8");
  window.eval(ui.slice(ui.indexOf("function esc("), ui.indexOf("/* ---------- 页面路由")));
  window.eval(fs.readFileSync(path.join(__dirname, "../js/fleet.js"), "utf8"));
  const document = window.document;
  const click = selector => window.fleetClick({ target: document.querySelector(selector) });
  return { window, document, store, reads, calls, click, notices };
}

test("directory shows safe public entries, pages by cursor, and posts only the join intent", async context => {
  const { window, document, reads, calls, click } = setup(context);
  await window.renderFleet();
  assert.deepEqual(reads, [["directory", 0]]);
  assert.equal(document.querySelector("#fleet-body img"), null);
  assert.ok(document.getElementById("fleet-pane").textContent.includes("<img src=x>"));
  await click('[data-fleet-act="next"]');
  assert.deepEqual(reads.at(-1), ["directory", 12]);
  assert.equal(document.querySelector('[data-fleet-act="next"]').disabled, true);
  await click('[data-fleet-act="previous"]');
  await click('[data-fleet-act="recruit"]');
  assert.deepEqual(calls, [["guild_join", { item_id: "00000002" }]]);
  await click('[data-fleet-tab="collection"]');
  assert.match(document.querySelector(".fleet-standing").textContent, /第 78 名/);
  assert.match(document.querySelector(".fleet-list").textContent, /25\.0%/);
  assert.match(document.querySelector(".fleet-list").textContent, /7 \/ 28 张整卡/);
  assert.equal(document.querySelector('[data-fleet-tab="collection"]').getAttribute("aria-pressed"), "true");
});

test("cooperative commands use server eligibility and wait for confirmed grants", async context => {
  const guild = { name: "波港", leader_id: 14, code: "0000000E", listed: false, members: [{ user_id: 14, name: "甲" }], boss_week: "2026-W48", boss_hp: 500, boss_max_hp: 24000,
    harbor: { score: 4, line_limit: 10, my_cargo: 1, lines: [{ id: "charts", name: "绘图", cost: "体力 5", done: 1, eligible: true }, { id: "lamps", name: "灯塔", cost: "结晶 2", done: 0, eligible: false }], beacons: [{ tier: 1, name: "浅湾", goal: 1, gold: 100, ore: 2, eligible: true }] } };
  const { window, document, click, calls } = setup(context, guild);
  await window.renderFleet();
  assert.equal(document.querySelector('[data-fleet-act="supply"][data-value="lamps"]').disabled, true);
  await click('[data-fleet-act="supply"][data-value="lamps"]');
  await click('[data-fleet-act="supply"][data-value="charts"]');
  await click('[data-fleet-act="claim"]');
  await click('[data-fleet-act="listing"]');
  assert.deepEqual(calls, [["guild_supply", { item_id: "charts" }], ["guild_supply_claim", { count: 1 }], ["guild_listing", { enabled: true }]]);
  assert.equal(guild.harbor.lines[0].done, 1);
  assert.equal(guild.harbor.beacons[0].claimed, undefined);
});

test("older rankings and late session responses cannot replace the current tab", async context => {
  const { window, document, click, store } = setup(context);
  await window.renderFleet();
  const immediate = store.api;
  let release;
  store.api = () => new Promise(done => { release = done; });
  const old = click('[data-fleet-tab="power"]');
  store.api = immediate;
  await click('[data-fleet-tab="collection"]');
  release({ board: "power", entries: [{ name: "过时榜单" }] });
  await old;
  assert.doesNotMatch(document.getElementById("fleet-pane").textContent, /过时榜单/);
  store.api = () => new Promise(done => { release = done; });
  const outgoing = click('[data-fleet-tab="power"]');
  store.sessionEpoch++;
  release({ board: "power", entries: [{ name: "其他账号" }] });
  await outgoing;
  assert.doesNotMatch(document.getElementById("fleet-pane").textContent, /其他账号/);
});
