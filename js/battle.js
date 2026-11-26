/* =====================================================
 * battle.js —— 播放服务器战报；胜负、消耗与掉落不在浏览器计算
 * ===================================================== */

let BAT = null;   // 当前战斗状态
let battleRunning = false;

function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }

/* ---------- 开战入口 ---------- */
async function startBattle(stageId) {
  if (!requireGame() || battleRunning) return;
  const epoch = Store.sessionEpoch;
  battleRunning = true;
  const run = await Store.startRun(stageId);
  if (epoch !== Store.sessionEpoch) return;
  if (!run) {
    battleRunning = false;
    if (document.body.dataset.page === "page-stage") updateStageRunLimits();
    if (Store.ready()) toast(Store.lastError || "服务器未能开启副本，请重新同步");
    return;
  }
  await resumeBattle(run, true);
}

async function resumeBattle(run, starting) {
  if (!requireGame() || (battleRunning && !starting)) return;
  let st = run.arena ? { id: "arena", name: "竞技场 · " + run.arena.name, type: "玩家对战", waves: 1, img: "assets/d06-observatory.webp" } : (typeof voyageStage === "function" && voyageStage(run.stage)) || STAGES.find(x => x.id === run.stage) || GACHA_RUN_MODES[run.stage];
  if (st && run.scene) {
    st = { ...st, name: run.scene.name, img: run.scene.img, starRound: run.scene.star_round };
  }
  const report = run.result;
  if (!st || !report || !Array.isArray(report.allies) || !Array.isArray(report.waves) ||
      !Array.isArray(report.events) || !Number.isFinite(run.ready_at) ||
      !Number.isFinite(report.duration_ms) || report.duration_ms <= 0) {
    battleRunning = false;
    return toast("服务器战报不完整，请重新同步");
  }
  st = { ...st, waves: report.waves.length };
  const allies = report.allies.map(u => ({ ...u, girl: GIRLS.find(g => g.id === u.girlId) }));
  const waves = report.waves.map(wave => wave.map(u => ({ ...u, girl: GIRLS.find(g => g.id === u.girlId) })));
  // HTTP Date has whole-second precision; do not add that rounding to a fresh report.
  const remaining = Math.min(report.duration_ms, Math.max(0, run.ready_at - Store.serverNow()));
  const battle = {
    stage: st, allies, waves, waveIdx: 0, log: [], rounds: 0, runId: run.id,
    report, arena: run.arena || null, readyAt: run.ready_at, readyDeadline: performance.now() + remaining,
    elapsed: Math.max(0, report.duration_ms - remaining), epoch: Store.sessionEpoch, abort: false,
  };
  BAT = battle;
  battleRunning = true;
  try {
    nav("page-battle");
    document.getElementById("scene-bg").style.backgroundImage = "url(" + st.img + ")";
    renderBattle();
    batLog("进入「" + st.name + "」，正在播放服务器战报");
    await playServerBattle(battle);
  } catch (e) {
    if (battleCurrent(battle)) {
      if (battle.arena) { finishArenaReplay(battle); return; }
      battle.playbackFailed = true;
      renderBattleControls(battle, "战报播放中断，可重新同步或重试结算");
      showBattleRecovery(battle, "战报播放中断，奖励仍由服务器结算，不会丢失或重复发放。");
    }
  } finally {
    if (BAT === battle) battleRunning = !battle.settled && !battle.abort;
  }
}

/* ---------- 渲染 ---------- */
const cutOf = src => src.replace(/\.webp$/, "-cut.webp");   // 透明立绘
function unitHTML(u, idx) {
  const hpPct = Math.max(0, u.hp / u.maxHp * 100);
  const face = cutOf(u.girl ? u.girl.img : u.img);
  const element = ELEM[u.elem];
  return `<div class="unit ${u.side === "enemy" ? "enemy" : ""} ${u.boss ? "boss" : ""} ${u.alive ? "" : "dead"}" id="u-${u.uid}" style="--bd:${-(idx * 0.9).toFixed(2)}s;--ed:${(idx * 0.09).toFixed(2)}s">
    <div class="u-body"><img class="u-face" src="${face}" alt=""></div>
    <div class="u-name">${esc(u.name)}${element ? `<span class="u-elem" style="--elem:${element.color}" title="${element.name}系">${element.name}</span>` : ""}</div>
    <div class="hp-bar"><div class="hp-fill" style="width:${hpPct}%"></div></div>
    <div class="mp-bar"><div class="mp-fill" style="width:${u.rage}%"></div></div>
    <div class="u-states" aria-label="状态">${unitStatusMarkup(u)}</div>
  </div>`;
}
function unitStatusMarkup(u) {
  return `${u.dotT > 0 ? `<span class="poison-state" title="每次行动受到侵蚀，治疗与吸血减半">侵蚀 ${u.dotT}</span>` : ""}${u.stunT > 0 ? '<span>眩晕</span>' : ""}${u.shield > 0 ? '<span class="shield-state">护盾</span>' : ""}${u.reflectT > 0 ? `<span class="shield-state">盾反${u.reflectPct}%</span>` : ""}${u.atkUp > 0 ? `<span>攻+${u.atkUp}%</span>` : ""}`;
}
function assignUids(list, side) { list.forEach((u, i) => u.uid = side + i); }

function renderBattle() {
  const scene = document.getElementById("battle-scene");
  const log = document.getElementById("battle-log");
  // Keep cached HTML shells compatible with the independent report panel.
  if (log.parentElement === scene) scene.after(log);
  const cur = BAT.waves[BAT.waveIdx];
  assignUids(BAT.allies, "a");
  assignUids(cur, "e");
  renderBattleHud();
  const boxes = [["side-ally", BAT.allies], ["side-enemy", cur]];
  boxes.forEach(([id, list]) => {
    const box = document.getElementById(id);
    box.innerHTML = list.map((u, i) => unitHTML(u, i)).join("");
    [...box.children].forEach(el => el.classList.add("enter"));   // 登场
  });
  setTimeout(() => {
    boxes.forEach(([id]) =>
      [...document.getElementById(id).children].forEach(el => el.classList.remove("enter")));
  }, 850);
}
function updateUnit(u) {
  const el = document.getElementById("u-" + u.uid);
  if (!el) return;
  el.querySelector(".hp-fill").style.width = Math.max(0, u.hp / u.maxHp * 100) + "%";
  el.querySelector(".mp-fill").style.width = u.rage + "%";
  el.classList.toggle("dead", !u.alive);
  el.classList.toggle("has-shield", (u.shield || 0) > 0);
  const states = el.querySelector(".u-states");
  if (states) states.innerHTML = unitStatusMarkup(u);
}
function batLog(msg) {
  BAT.log.push(msg);
  const el = document.querySelector("#battle-log .bl-lines");
  el.innerHTML = BAT.log.slice(-40).map(m => "<p>" + m + "</p>").join("");
  el.scrollTop = el.scrollHeight;
}
/* 顶部战斗 HUD：关卡名、类型、波次点 */
function renderBattleHud() {
  const st = BAT.stage;
  document.getElementById("bh-name").textContent = st.name;
  document.getElementById("bh-type").textContent = st.type;
  const box = document.getElementById("bh-waves");
  box.innerHTML = Array.from({ length: st.waves }, (_, i) =>
    `<i class="${i <= BAT.waveIdx ? "on" : ""}"></i>`).join("");
  renderBattleControls(BAT, "正在播放服务器战报");
}
function renderBattleControls(battle, message) {
  const button = document.getElementById("btn-flee");
  let status = document.getElementById("bh-status");
  // An older cached HTML shell can load this script during a rolling release.
  if (!status && button) {
    status = document.createElement("span");
    status.id = "bh-status";
    status.setAttribute("role", "status");
    status.setAttribute("aria-live", "polite");
    button.before(status);
  }
  if (status) status.textContent = message;
  if (!button) return;
  button.onclick = battleControl;
  button.disabled = Boolean(battle.settling);
  if (battle.arena) {
    button.textContent = battle.settled ? "返回竞技场" : "跳过战报";
    return;
  }
  if (battle.settled) button.textContent = battle.stage.progression ? "返回成长" : "返回副本";
  else if (battle.playbackFailed) button.textContent = "同步战报";
  else if (battle.playbackComplete) button.textContent = "重新同步";
  else button.textContent = "撤退";
}
function floatText(u, text, cls) {
  const el = document.getElementById("u-" + u.uid);
  if (!el) return;
  const t = document.createElement("div");
  t.className = "float-text " + cls;
  t.textContent = text;
  el.appendChild(t);
  setTimeout(() => t.remove(), 900);
}

/* ---------- 战斗特效 ---------- */
const FX_COLOR = {   // 每位少女的技能特效色
  c01: "#ff8a5c", c02: "#9adf6e", c03: "#6ec6ff", c04: "#ff6e5e", c05: "#a8e0ff",
  c06: "#a5e07a", c07: "#ffe27a", c08: "#ffb0c8", c09: "#c9a2ff", c10: "#b06eff",
  c11: "#ff9c5c", c12: "#8ff0b6", c13: "#8fd0ff", c14: "#9a7aff",
};
function unitPoint(u) {
  const el = document.getElementById("u-" + u.uid);
  if (!el) return null;
  const s = document.getElementById("battle-scene").getBoundingClientRect();
  const r = el.getBoundingClientRect();
  return { x: r.left - s.left + r.width / 2, y: r.top - s.top + r.height * 0.35 };
}
function spawnFx(x, y, cls, color) {
  const d = document.createElement("div");
  d.className = "fx " + cls;
  if (color) d.style.setProperty("--fx-c", color);
  d.style.left = x + "px"; d.style.top = y + "px";
  document.getElementById("fx-layer").appendChild(d);
  setTimeout(() => d.remove(), 750);
}
function hitAnim(u) {
  const el = document.getElementById("u-" + u.uid);
  if (!el) return;
  const dir = u.side === "ally" ? "hit-r" : "hit-l";   // 沿受击方向击退
  el.classList.remove("hit-r", "hit-l"); void el.offsetWidth;
  el.classList.add(dir);
  setTimeout(() => el.classList.remove(dir), 380);
}
function lungeAnim(u) {
  const el = document.getElementById("u-" + u.uid);
  if (!el) return;
  el.classList.remove("attacking"); void el.offsetWidth; el.classList.add("attacking");
  ghostTrail(u);   // 冲刺残影
  const p = unitPoint(u);
  if (p) spawnFx(p.x + (u.side === "ally" ? -22 : 22), p.y, "fx-dash");
  setTimeout(() => el.classList.remove("attacking"), 470);
}
function leanAnim(u) {   // 远程：拉弓后仰
  const el = document.getElementById("u-" + u.uid);
  if (!el) return;
  el.classList.remove("lean"); void el.offsetWidth; el.classList.add("lean");
  setTimeout(() => el.classList.remove("lean"), 400);
}
function castAnim(u) {   // 法术：浮空施法
  const el = document.getElementById("u-" + u.uid);
  if (!el) return;
  el.classList.remove("casting"); void el.offsetWidth; el.classList.add("casting");
  setTimeout(() => el.classList.remove("casting"), 640);
}
function jumpAnim(u) {   // 重击：跳起下砸
  const el = document.getElementById("u-" + u.uid);
  if (!el) return;
  el.classList.remove("attacking-jump"); void el.offsetWidth; el.classList.add("attacking-jump");
  setTimeout(() => el.classList.remove("attacking-jump"), 620);
}
function ghostTrail(u) { // 冲刺残影：复制立绘渐隐
  const el = document.getElementById("u-" + u.uid);
  const face = el && el.querySelector(".u-face");
  if (!el || !face) return;
  const s = document.getElementById("battle-scene").getBoundingClientRect();
  const r = face.getBoundingClientRect();
  const x = r.left - s.left + r.width / 2, y = r.top - s.top + r.height / 2;
  const dir = u.side === "ally" ? 1 : -1;   // 攻击方向
  for (let i = 1; i <= 2; i++) {
    const g = document.createElement("div");
    g.className = "fx fx-ghost";
    g.style.setProperty("--g-img", 'url("' + face.src + '")');
    g.style.setProperty("--g-dx", (-dir * 18 * i) + "px");
    g.style.left = x + "px"; g.style.top = y + "px";
    g.style.width = r.width + "px"; g.style.height = r.height + "px";
    g.style.animationDelay = (i * 45) + "ms";
    document.getElementById("fx-layer").appendChild(g);
    setTimeout(() => g.remove(), 650);
  }
}
function projectileAnim(src, dst, color) {  // 远程弹道
  return new Promise(res => {
    const sp = unitPoint(src), dp = unitPoint(dst);
    if (!sp || !dp) return res();
    const orb = document.createElement("div");
    orb.className = "fx fx-orb";
    if (color) orb.style.setProperty("--fx-c", color);
    orb.style.left = sp.x + "px"; orb.style.top = sp.y + "px";
    document.getElementById("fx-layer").appendChild(orb);
    requestAnimationFrame(() => requestAnimationFrame(() => {
      orb.style.transform = "translate(" + (dp.x - sp.x) + "px," + (dp.y - sp.y) + "px) translate(-50%,-50%)";
    }));
    setTimeout(() => {
      orb.remove();
      spawnFx(dp.x, dp.y, "fx-burst", color);
      res();
    }, 210);
  });
}
function shieldAnim(u) {
  const el = document.getElementById("u-" + u.uid);
  if (!el) return;
  el.classList.remove("shielded"); void el.offsetWidth; el.classList.add("shielded");
  setTimeout(() => el.classList.remove("shielded"), 550);
}
function skillBanner(name) {
  const b = document.getElementById("skill-banner");
  b.innerHTML = '<div class="sk-name">' + name + "</div>";
  b.classList.remove("hidden"); void b.offsetWidth;
  clearTimeout(b._t);
  b._t = setTimeout(() => b.classList.add("hidden"), 750);
}
function sceneShake() {
  const s = document.getElementById("battle-scene");
  s.classList.remove("shake"); void s.offsetWidth; s.classList.add("shake");
  setTimeout(() => s.classList.remove("shake"), 420);
}
function flashScene() {
  const f = document.getElementById("scene-flash");
  f.classList.remove("hidden", "go"); void f.offsetWidth;
  f.classList.add("go");
  setTimeout(() => f.classList.add("hidden"), 520);
}
function waveTitle(text) {
  const w = document.getElementById("wave-title");
  w.textContent = text;
  w.classList.remove("hidden", "go"); void w.offsetWidth;
  w.classList.add("go");
  setTimeout(() => w.classList.add("hidden"), 1150);
}
/* 一次攻击的完整演出：按攻击类型播放动作 -> 命中特效（受击反馈来自服务器战报） */
function atkTypeOf(u) {
  if (u.girl) return u.girl.atkType || "melee";
  return u.atkType || "melee";
}
async function playStrike(src, dst, color) {
  const t = atkTypeOf(src);
  if (t === "magic") {
    castAnim(src);
    await sleep(260);
    const p = unitPoint(dst);
    if (p) spawnFx(p.x, p.y, "fx-burst", color);
  } else if (t === "ranged") {
    leanAnim(src);
    await sleep(140);
    await projectileAnim(src, dst, color);
  } else if (t === "heavy") {
    jumpAnim(src);
    await sleep(340);                     // 跳到下砸落地帧
    const p = unitPoint(dst);
    if (p) spawnFx(p.x, p.y + 14, "fx-burst", color);
  } else {                                // 近战：蓄力突进斩
    lungeAnim(src);
    await sleep(190);
    const p = unitPoint(dst);
    if (p) spawnFx(p.x, p.y, "fx-slash", color);
  }
  await sleep(70);
}

/* ---------- 服务器战报播放 ---------- */
function battleCurrent(battle) {
  return Boolean(battle && BAT === battle && !battle.abort && battle.epoch === Store.sessionEpoch && Store.ready());
}
function battleUnit(battle, uid) {
  return [...battle.allies, ...battle.waves[battle.waveIdx]].find(u => u.uid === uid);
}
function mergeBattleEvent(battle, event) {
  for (const snapshot of event.units || []) {
    const unit = battleUnit(battle, snapshot.uid);
    if (unit) { Object.assign(unit, snapshot); updateUnit(unit); }
  }
  battle.rounds = event.round;
}
async function playServerBattle(battle) {
  // 刷新页面后跳过已播放的时间段；这里只重放服务器快照，不计算战斗。
  const elapsed = battle.elapsed;
  let timeline = 0;
  for (const event of battle.report.events) {
    if (!battleCurrent(battle)) return;
    const duration = Math.max(0, event.duration_ms || 0);
    const skipped = timeline + duration <= elapsed;
    timeline += duration;
    if (event.kind === "wave") {
      battle.waveIdx = event.wave;
      renderBattle();
      mergeBattleEvent(battle, event);
      if (!skipped) {
        flashScene();
        const boss = battle.stage.hasBoss && event.wave === battle.waves.length - 1;
        waveTitle(boss ? "BOSS 来袭" : "第 " + (event.wave + 1) + " 波");
        if (boss) sceneShake();
        batLog("—— 第 " + (event.wave + 1) + " 波 ——");
        await sleep(duration);
      }
      continue;
    }
    if (skipped) { mergeBattleEvent(battle, event); continue; }
    const started = performance.now();
    const src = battleUnit(battle, event.src);
    if (src) {
      document.getElementById("u-" + src.uid)?.classList.add("acting");
      const color = src.side === "ally" ? (FX_COLOR[src.girlId] || "#ffd76e") : "#ff8a8a";
      if (event.kind === "ailment") batLog(esc(src.name) + " 受到侵蚀伤害");
      else if (event.skipped) batLog(src.name + " 眩晕中，无法行动");
      else if (event.skill) {
        castAnim(src); skillBanner(event.skill_name);
        batLog(src.name + " 发动【" + event.skill_name + "】");
        await sleep(300);
      } else {
        const target = (event.targets || []).find(t => t.damage > 0 || t.absorbed > 0);
        const dst = target && battleUnit(battle, target.uid);
        if (dst) { batLog(src.name + " 攻击 " + dst.name); await playStrike(src, dst, color); }
      }
      if (!battleCurrent(battle)) return;
      for (const target of event.targets || []) {
        const dst = battleUnit(battle, target.uid);
        if (!dst) continue;
        const pos = unitPoint(dst);
        if (target.elemental) {
          const label = target.elemental === "advantage" ? "属性克制 +20%" : "属性受克 −20%";
          batLog(`${esc(src.name)} 对 ${esc(dst.name)} ${label}`);
        }
        if (target.damage > 0 || target.absorbed > 0) {
          hitAnim(dst);
          if (target.damage > 0) floatText(dst, "-" + target.damage, target.crit ? "crit dmg" : "dmg");
          if (target.absorbed > 0) floatText(dst, "盾" + target.absorbed, "heal");
          if (pos && event.skill) spawnFx(pos.x, pos.y, "fx-burst", color);
        }
        if (target.heal > 0) {
          floatText(dst, "+" + target.heal, "heal");
          if (pos) spawnFx(pos.x, pos.y, "fx-heal");
        }
        if (target.shield > 0) {
          shieldAnim(dst); floatText(dst, "盾+" + target.shield, "heal");
          if (pos) spawnFx(pos.x, pos.y, "fx-ring", "#6ea0ff");
        }
        if (target.self_heal > 0) {
          floatText(src, "+" + target.self_heal, "heal");
          batLog(`${esc(src.name)} 吸血恢复 ${target.self_heal}`);
        }
        if (target.reflected > 0) {
          floatText(src, "反伤-" + target.reflected, "dmg");
          batLog(`${esc(src.name)} 受到反击/映伤 ${target.reflected}`);
        }
        if (target.rage_delta < 0) {
          floatText(dst, "怒气" + target.rage_delta, "dmg");
          batLog(`${esc(dst.name)} 被封怒，怒气 ${target.rage_delta}`);
        }
        if (target.rage_delta > 0) {
          floatText(dst, "怒气+" + target.rage_delta, "heal");
          batLog(`${esc(dst.name)} 获得怒气 +${target.rage_delta}`);
        }
      }
      document.getElementById("u-" + src.uid)?.classList.remove("acting");
    }
    mergeBattleEvent(battle, event);
    await sleep(Math.max(0, duration - (performance.now() - started)));
  }
  battle.playbackComplete = true;
  while (battleCurrent(battle) && performance.now() < battle.readyDeadline) {
    const remaining = battle.readyDeadline - performance.now();
    renderBattleControls(battle, "战报已结束，" + Math.ceil(remaining / 1000) + "秒后自动结算");
    await sleep(Math.min(1000, remaining));
  }
  if (battleCurrent(battle)) await battleEnd(battle);
}

/* ---------- 数据库结算 ---------- */
function showBattleRecovery(battle, message) {
  if (!battleCurrent(battle)) return;
  showModal(`<h3>结算尚未确认</h3>
    <p style="color:var(--sub);text-align:center">${esc(message)}<br>产出以数据库为准，重试不会重复发奖。</p>
    <div class="modal-btns">
      <button class="gbtn gbtn-sm gbtn-primary" onclick="closeModal(true);battleEnd(BAT)"><span>重试结算</span></button>
      <button class="gbtn gbtn-sm" onclick="closeModal(true);recoverBattle()"><span>重新同步</span></button>
      <button class="gbtn gbtn-sm" onclick="closeModal(true);exitBattle()"><span>放弃本局</span></button>
    </div>`);
}
function battleDropLedger(drops) {
  const items = Object.entries(Store.profile?.items?.items || {});
  return drops.map(line => {
    const match = typeof line === "string" && /^(.+?)\s+[x×]\s*(\d+)$/.exec(line);
    const entry = match && items.find(([, item]) => item.name === match[1]);
    if (!entry || typeof atelierIcon !== "function") return `<li>${esc(line)}</li>`;
    return `<li class="battle-item-drop">${atelierIcon(entry[0])}<b>${esc(entry[1].name)}</b><strong>×${match[2]}</strong></li>`;
  }).join("") || "<li>无掉落</li>";
}

async function battleEnd(battle) {
  if (!battleCurrent(battle) || battle.settling || battle.settled) return;
  if (battle.arena) { finishArenaReplay(battle); return; }
  battle.settling = true;
  renderBattleControls(battle, "正在结算…");
  const result = await Store.finishRun(battle.runId);
  battle.settling = false;
  if (!battleCurrent(battle)) return;
  if (!result) {
    renderBattleControls(battle, "结算未确认，请重试或重新同步");
    showBattleRecovery(battle, Store.lastError || "网络异常");
    return;
  }
  battle.settled = true;
  battleRunning = false;
  renderBattleControls(battle, "已结算");
  const win = result.win;
  const st = battle.stage;
  const drops = [...(result.drops || [])];
  batLog(win ? "战斗胜利！数据库结算完成" : "战斗失败，无掉落");
  waveTitle(win ? "胜 利" : "战 斗 失 败");
  if (win) {
    flashScene();
    battle.allies.forEach(u => {
      if (!u.alive) return;
      document.getElementById("u-" + u.uid)?.classList.add("victory");
      const p = unitPoint(u);
      if (p) spawnFx(p.x, p.y, "fx-ring", "#ffcf6e");
    });
  }
  // 余额同样重新读取，绝不在客户端把结算额度累加。
  Store.refresh().then(d => { if (d && battleCurrent(battle)) renderTopbar(); });
  renderTopbar();
  if (!battleCurrent(battle)) return;
  showModal(`
    <h3>${win ? "战斗胜利" : "战斗失败"}</h3>
    ${win ? `<div class="res-stars">${starsHTML(result.stars, 26)}</div>
    <div class="res-cond">
      <span class="${result.alive_n === result.total_n ? "ok" : ""}">全员存活 ${result.alive_n}/${result.total_n}</span>
      <span class="${result.fast ? "ok" : ""}">${result.rounds} 回合 · 目标 ${st.starRound}</span>
      ${result.is_best ? `<span class="new">刷新最佳</span>` : ""}
    </div>` : ""}
    <p style="color:var(--sub);font-size:12px;text-align:center">${esc(st.name)} · ${esc(st.type)}</p>
    ${typeof voyageResultSummary === "function" ? voyageResultSummary(st.id, result) : ""}
    <ul class="drop-list">${battleDropLedger(drops)}</ul>
    <div class="modal-btns">
      <button class="gbtn gbtn-sm" onclick="closeModal();nav('${battleDestination(st)}')"><span>${st.progression ? "返回成长" : GACHA_RUN_MODES[st.id] ? "返回玩法" : "返回副本"}</span></button>
      ${GACHA_RUN_MODES[st.id] && !st.progression ? "" : `<button class="gbtn gbtn-sm gbtn-primary" onclick="closeModal();retryBattle('${st.id}')"><span>再来一次</span></button>`}
    </div>`);
}
function retryBattle(id) {
  if (!requireGame()) return;
  startBattle(id);
}
function battleDestination(stage) {
  if (stage.progression) return "page-growth";
  return GACHA_RUN_MODES[stage.id] ? "page-modes" : "page-stage";
}
async function recoverBattle() {
  const battle = BAT;
  if (!battleCurrent(battle) || battle.settling) return;
  if (battle.arena) { finishArenaReplay(battle); return; }
  const profile = await Store.refresh();
  if (!profile || !battleCurrent(battle) || battle.settling || battle.settled) return;
  battle.abort = true;
  BAT = null;
  battleRunning = false;
  nav(battleDestination(battle.stage));
  if (Store.activeRun) resumeBattle(Store.activeRun);
  else toast("已同步，本局已结算或过期");
}
function battleControl() {
  if (BAT?.arena) {
    if (BAT.settled) leaveArenaReplay();
    else { BAT.abort = true; battleRunning = false; showArenaResult(BAT.arena); }
    return;
  }
  if (BAT && (BAT.playbackComplete || BAT.playbackFailed) && !BAT.settled) return recoverBattle();
  return exitBattle();
}
/* 撤退也要由服务器确认，网络失败可用同一run_id安全重试。 */
async function exitBattle() {
  const battle = BAT;
  if (!battleCurrent(battle) || battle.settling) return;
  if (battle.arena) { leaveArenaReplay(); return; }
  if (battle.settled) {
    battle.abort = true;
    BAT = null;
    battleRunning = false;
    closeModal(true);
    nav(battleDestination(battle.stage));
    return;
  }
  battle.settling = true;
  const result = await Store.finishRun(battle.runId, true);
  battle.settling = false;
  if (!battleCurrent(battle)) return;
  if (!result) {
    toast("撤退未确认，请重试：" + (Store.lastError || "网络异常"));
    if (battle.playbackComplete) await battleEnd(battle);
    return;
  }
  battle.abort = true;
  battleRunning = false;
  BAT = null;
  nav(battleDestination(battle.stage));
  toast(battle.stage.progression && !battle.stage.cost ? "已撤退，未发奖励，也未消耗训练资格" : "已撤退，体力不返还");
}

function finishArenaReplay(battle) {
  if (!battleCurrent(battle)) return;
  battle.settled = true;
  battleRunning = false;
  renderBattleControls(battle, "挑战已结算");
  showArenaResult(battle.arena);
}
