const VOYAGE_MODES = { story: "剧情", hard: "困难", nightmare: "噩梦" };
const VOYAGE_PORTS = { story: "主线航图", resources: "资源航线", rifts: "元素裂隙" };
let voyageView = { epoch: -1, port: "story", chapter: 0, mode: "story", rift: 0 };
let voyageNoticeTimer;

function voyageSession() {
  if (voyageView.epoch !== Store.sessionEpoch) voyageView = { epoch: Store.sessionEpoch, port: "story", chapter: 0, mode: "story", rift: 0 };
  return voyageView;
}

function openVoyage(port, stageId) {
  if (!requireGame() || !Object.hasOwn(VOYAGE_PORTS, port)) return;
  if (stageId) closeModal();
  voyageSession().port = port;
  if (!Store.profile?.voyage && port === "rifts") stageView.ch = 4;
  nav("page-stage");
  if (stageId && voyageContext()?.ids.includes(stageId)) voyageDetail(stageId);
  if (!Store.profile?.voyage && port === "resources") toast("当前服务器尚未开放定向资源航线");
}

function voyageStage(id) {
  const local = STAGES.find(stage => stage.id === id);
  const point = Store.profile?.voyage?.stages?.[id];
  if (!point) return local;
  return { ...local, ...point, type: VOYAGE_MODES[point.kind] || "资源航线", desc: point.description || local?.desc || "" };
}

function voyageGate(point) {
  const reasons = [];
  if ((point.captain_min || 0) > (SAVE.captain?.level || 1)) reasons.push(`船长达到 Lv.${point.captain_min}`);
  for (const id of point.requires || (point.unlockStage ? [point.unlockStage] : [])) {
    if (!SAVE.cleared[id]) reasons.push(`通关「${voyageStage(id)?.name || id}」`);
  }
  if (point.chapter_gate && !SAVE.voyage_passes?.[point.id] && !SAVE.cleared[point.id] && !SAVE.cleared[point.chapter_gate]) reasons.push(`完成前章「${voyageStage(point.chapter_gate)?.name || point.chapter_gate}」`);
  return reasons.join("、");
}

function voyageContext() {
  const view = voyageSession(), book = Store.profile?.voyage;
  if (!book?.chapters?.length || !book.stages || !book.resources?.length) return null;
  const chapter = book.chapters[view.chapter] || book.chapters[0];
  if (view.port === "resources") return { view, book, chapter, ids: book.resources, title: "群岛补给线", caption: "定向物资 · 每港每日三次", chests: [], region: 2 };
  if (view.port === "rifts") {
    const rift = CHAPTERS[4 + view.rift];
    return { view, book, chapter, ids: rift.stages, title: rift.name, caption: "灵印打捞 · 六系守门人", chests: [], region: 3 };
  }
  return { view, book, chapter, ids: chapter.routes[view.mode], title: chapter.name, caption: `第 ${view.chapter + 1} 海域 · ${chapter.landmark}`, chests: chapter.chests[view.mode], region: view.chapter };
}

function voyageCoordinates(count, portrait) {
  if (count === 2) return portrait ? [[28,76],[72,22]] : [[26,70],[74,27]];
  if (count === 3) return portrait ? [[27,80],[74,49],[27,17]] : [[17,72],[49,24],[82,64]];
  if (portrait) {
    const rows = Math.ceil(count / 2);
    return Array.from({ length: count }, (_, index) => {
      const row = Math.floor(index / 2), left = (row + index % 2) % 2 === 0;
      return [left ? 28 : 73, 84 - row * (68 / (rows - 1))];
    });
  }
  const positions = count === 6 ? [[15,72],[42,76],[72,64],[85,26],[52,21],[21,32]] : [[12,74],[32,60],[52,74],[72,60],[86,29],[65,19],[43,34],[22,19]];
  return positions.slice(0, count);
}

function voyageChart(context) {
  const query = "(max-width: 650px) and (orientation: portrait)";
  const portrait = window.matchMedia ? window.matchMedia(query).matches : window.innerWidth < 650;
  const coords = voyageCoordinates(context.ids.length, portrait), width = portrait ? 400 : 1000, height = portrait ? 620 : 480;
  const points = coords.map(([left, top]) => [left * width / 100, top * height / 100]);
  const next = context.ids.find(id => !SAVE.cleared[id] && !voyageGate(voyageStage(id)));
  const routes = points.slice(1).map(([endX, endY], index) => {
    const [startX, startY] = points[index], midX = (endX + startX) / 2, midY = (endY + startY) / 2;
    const path = Math.abs(startX - endX) < 10 ? `M${startX},${startY} Q${midX + 45},${midY} ${endX},${endY}` : `M${startX},${startY} C${midX},${startY + 38} ${midX},${endY - 42} ${endX},${endY}`;
    return `<path d="${path}" class="${SAVE.cleared[context.ids[index + 1]] ? "sailed" : ""}"/>`;
  }).join("");
  const islands = points.map(([left, top], index) => `<g transform="translate(${left} ${top + 7}) scale(${portrait ? .75 : 1} ${portrait ? .79 : 1})"><use href="#voy-islet" transform="scale(${index % 2 ? -.95 : 1.02} ${index % 3 ? .97 : 1.1})"/><use href="#voy-peaks" x="${index % 2 ? 12 : -12}"/></g>`).join("");
  const nodes = context.ids.map((id, index) => {
    const point = voyageStage(id), gate = voyageGate(point), stars = Math.max(0, Math.min(3, Number(SAVE.stars[id]) || 0));
    const [left, top] = coords[index], state = gate ? "未解锁" : SAVE.cleared[id] ? "已通关" : "可挑战";
    return `<button type="button" data-voy-point="${esc(id)}" style="--vx:${left}%;--vy:${top}%" class="voy-node ${gate ? "is-locked" : ""} ${id === next ? "is-next" : ""}" title="${esc(gate || state)}" aria-label="${esc(point.name)}，${state}，${stars}星"><span class="voy-node-marker" aria-hidden="true">${gate ? "◇" : point.hasBoss ? "♜" : id === next ? "⚑" : SAVE.cleared[id] ? "✓" : index + 1}</span><b>${esc(point.name)}</b><small aria-hidden="true">${context.view.port === "resources" ? esc(voyageItem(point.prize[0].item)) : `${"★".repeat(stars)}${"☆".repeat(3-stars)}`}</small></button>`;
  }).join("");
  return `<svg class="voy-waters" viewBox="0 0 ${width} ${height}" preserveAspectRatio="none" aria-hidden="true"><defs><g id="voy-islet"><path class="shelf" d="M-72 12l8-18 23-5 11-16 29 6 15-9 21 13 19 3 15 18-13 18-33 9-21-4-20 7-24-9z"/><path class="inland" d="M-57 10l12-10 19-8 8-11 26 9 21-6 17 14 14 10-26 13-24-4-19 4-21-6z"/></g><g id="voy-peaks"><path d="m-30 4 11-20L-9 3m2-1 15-29 18 30m-18-30 1 19 9-1"/></g></defs><g class="voy-route-lines">${routes}</g>${islands}</svg>${nodes}`;
}

function voyageLandmark(region) {
  const silhouettes = [
    "M8 38h42M20 35l3-17h11l4 17M22 17h13v-6H22zM18 11l11-8 10 8M26 13v3m5-3v3M5 39l8-5 6 3m20-2 7-3 8 6",
    "M4 36h48M12 34V17L33 6l5 2-18 15v11m1-11h8M35 10v15c0 4 5 4 5 0M29 32l5 1 6-2 7 1-4 5H32zM5 40q6-3 12 0t12 0t12 0",
    "M8 34q20 10 40-1l-5 7H15zM26 33V4m-2 4-14 21 14-2zm5 1 12 16-12 2zM5 43q8-4 16-1t18-1",
    "M8 36h42M17 35l6-8m11 0 6 8M30 9a10 10 0 1 1-1 0M12 23c-3-5 25-19 32-14s-19 20-28 18M29 7V3m13 28 3 4M7 13l2-3 2 3-2 3z",
  ];
  return `<svg aria-hidden="true" viewBox="0 0 58 46"><path d="${silhouettes[region] || silhouettes[0]}"/></svg>`;
}

function voyageItem(id) {
  return id === "gold" ? "金币" : Store.profile?.items?.items?.[id]?.name || id;
}

function voyagePrizes(prizes) {
  return (prizes || []).map(prize => `<span class="voy-prize">${esc(voyageItem(prize.item))} <b>×${esc(prize.amount)}</b></span>`).join("");
}

function voyageTotal(ids, state = SAVE) {
  return ids.reduce((sum, id) => sum + Math.max(0, Math.min(3, Number(state.stars[id]) || 0)), 0);
}

function renderVoyage() {
  const mount = document.getElementById("voyage-map"), context = voyageContext();
  if (!mount || !context) {
    delete document.body.dataset.voyage;
    if (mount) { mount.hidden = true; mount.replaceChildren(); }
    return false;
  }
  const { view, chapter, chests, ids, book, title, caption } = context;
  const total = voyageTotal(ids), story = view.port === "story", maxStars = 3 * ids.length;
  const tabs = story ? Object.entries(VOYAGE_MODES).map(([id, label]) => {
    const gate = voyageGate(voyageStage(chapter.routes[id][0]));
    return `<button type="button" data-voy-mode="${id}" aria-pressed="${id === view.mode}" title="${esc(gate ? `解锁条件：${gate}` : label)}"><i aria-hidden="true">${{story:"✧",hard:"⚔",nightmare:"☾"}[id]}</i>${label}${gate ? '<small aria-label="未解锁">◇</small>' : ""}</button>`;
  }).join("") : view.port === "rifts" ? ["普通裂隙","噩梦回响"].map((label, index) => `<button type="button" data-voy-rift="${index}" aria-pressed="${view.rift === index}">${label}</button>`).join("") : '<span class="voy-port-label">定向补给 · 北京零点换日</span>';
  const chestPanel = story ? `<section class="voy-chests" aria-label="章节星级宝箱"><div><strong>航路星箱</strong><span>${total} / ${maxStars} ★</span></div><nav>${chests.map(chest => {
    const claimed = !!SAVE.voyage_claims?.[chest.id], ready = total >= chest.stars && !claimed;
    return `<button type="button" data-voy-chest="${esc(chest.id)}" class="${ready ? "is-ready" : ""}" aria-label="${chest.stars}星箱，${claimed ? "已领取" : ready ? "可领取" : `还差${chest.stars - total}星`}"><span aria-hidden="true">${claimed ? "✓" : "▣"}</span>${chest.stars}★<small>${claimed ? "已领" : ready ? "可领" : "未满"}</small></button>`;
  }).join("")}</nav><progress value="${total}" max="${maxStars}" aria-label="本章此难度星数"></progress></section>` : `<section class="voy-legend"><strong>${view.port === "resources" ? "补给清单" : "裂隙图谱"}</strong><span>${view.port === "resources" ? "每港独立计次 · 胜利取得指定物资" : `已突破 ${ids.filter(id=>SAVE.cleared[id]).length} / ${ids.length} 道裂隙`}</span></section>`;
  const arrows = story ? `<nav class="voy-arrows" aria-label="切换章节">${view.chapter + 1 < book.chapters.length ? `<button type="button" data-voy-step="1" title="下一章 · ${esc(book.chapters[view.chapter+1].name)}" aria-label="下一章">›</button>` : ""}${view.chapter ? '<button type="button" data-voy-step="-1" title="上一章" aria-label="上一章">‹</button>' : ""}</nav>` : "";
  document.body.dataset.voyage = "open";
  mount.hidden = false;
  mount.dataset.region = context.region;
  mount.innerHTML = `<div class="voy-map-frame"><header class="voy-head"><button class="voy-back" type="button" data-voy-home aria-label="返回星霜港" title="返回星霜港">←</button><div class="voy-heading"><small>${esc(caption)}</small><h2>${esc(title)}</h2></div><div class="voy-wallet"><span>金币 <b>${esc(SAVE.gold)}</b></span><span>代币 <b>${esc(SAVE.token)}</b></span><button type="button" data-voy-stamina title="前往船长日志领取补给">体力 <b>${esc(SAVE.stamina)}/${esc(SAVE.staminaMax)}</b> ＋</button></div></header><nav class="voy-port-nav" aria-label="航图分类">${Object.entries(VOYAGE_PORTS).map(([id,label]) => `<button type="button" data-voy-port="${id}" aria-pressed="${view.port === id}">${label}</button>`).join("")}</nav><div class="voy-chart" role="group" aria-label="${esc(title)}航点"><span class="voy-compass" aria-hidden="true">✥<small>N</small></span><span class="voy-landmark" aria-hidden="true">${esc(story ? chapter.landmark : title)}</span>${voyageChart(context)}${arrows}</div><footer class="voy-foot"><nav class="voy-difficulties" aria-label="关卡难度">${tabs}</nav>${chestPanel}</footer><span class="voy-status" role="status"></span></div>`;
  mount.onclick = voyageClickMap;
  mount.querySelector(".voy-landmark")?.insertAdjacentHTML("afterbegin", voyageLandmark(context.region));
  updateStageRunLimits();
  return true;
}

function voyageClickMap(event) {
  const button = event.target.closest("button");
  if (!button || !requireGame()) return;
  const data = button.dataset, context = voyageContext();
  if ("voyHome" in data) return nav("page-home");
  if ("voyStamina" in data) return nav("page-captain");
  if (data.voyPort) return openVoyage(data.voyPort);
  if (data.voyPoint) return voyageDetail(data.voyPoint);
  if (data.voyChest) return voyageChest(data.voyChest);
  if (data.voyMode && Object.hasOwn(VOYAGE_MODES, data.voyMode)) context.view.mode = data.voyMode;
  if (data.voyRift === "0" || data.voyRift === "1") context.view.rift = Number(data.voyRift);
  if (data.voyStep) {
    const to = context.view.chapter + Number(data.voyStep), next = context.book.chapters[to];
    if (!next) return;
    const gate = voyageGate(voyageStage(next.routes[context.view.mode][0]));
    if (gate && to > context.view.chapter) {
      const notice = document.querySelector("#voyage-map .voy-status");
      notice.textContent = `下一章需${gate}`;
      clearTimeout(voyageNoticeTimer);
      voyageNoticeTimer = setTimeout(() => { notice.textContent = ""; }, 3500);
      return;
    }
    context.view.chapter = to;
  }
  renderVoyage();
  const group = ["voyMode", "voyRift", "voyStep"].find(key => key in data);
  if (group) Array.from(document.querySelectorAll("#voyage-map button")).find(candidate => candidate.dataset[group] === data[group])?.focus();
}

function voyageRemaining(point) {
  return point.daily_limit ? Math.max(0, point.daily_limit - (SAVE.daily.voyage_attempts?.[point.id] || 0)) : null;
}

function voyageDetail(id) {
  const point = voyageStage(id), context = voyageContext();
  if (!point || !context.ids.includes(id)) return;
  const gate = voyageGate(point), left = voyageRemaining(point), best = SAVE.stars[id] || 0, date = SAVE.daily.date;
  const milestones = context.view.port === "story" ? `<div class="voy-milestones"><p>首胜${SAVE.cleared[id] ? " · 已达成" : ""} ${voyagePrizes(context.book.first_clear)}</p><p>首次三星${best === 3 ? " · 已达成" : ""} ${voyagePrizes(context.book.first_three)}</p></div>` : "";
  const mobs = Array.from({length: Math.min(point.waves, 8)}, (_, wave) => {
    const boss = point.hasBoss && wave === point.waves-1;
    const names = boss ? [ENEMY_TPL[ENEMY_TPL.length - 1].name] : Array.from({length:point.enemyCount}, (_, enemy) => ENEMY_TPL[(wave * point.enemyCount + enemy) % 3].name);
    return `<p>第 ${wave+1} 波 · ${names.map(esc).join("、")}</p>`;
  }).join("");
  const rewards = point.prize ? `<div class="voy-drop-grid">${voyagePrizes(point.prize)}</div><p>出战伙伴经验 +${point.exp}</p>` : `<div class="voy-legacy-loot">${legacyStageDetailHTML(point, true)}</div>`;
  const sweep = point.sweep_prize?.length ? `<section class="voy-sweep"><h4>三星扫荡</h4><div class="voy-drop-grid">${voyagePrizes(point.sweep_prize)}</div><p>每次只发上述固定物资，消耗 ${point.cost} 体力${point.daily_limit ? "及一次挑战次数" : ""}。不发站点额度、首胜、首次三星或战斗奖励，不计战斗任务。</p><div class="voy-sweep-buttons">${[1,3].map(count => `<button type="button" class="gbtn gbtn-sm" data-voy-sweep="${count}"><span>扫荡 ${count} 次 · ${point.cost*count} 体力</span></button>`).join("")}</div><small data-voy-sweep-status></small></section>` : "";
  const mechanic = DUNGEON_MECHANICS[point.mechanic];
  showModal(`<h3>${esc(point.name)}</h3><article class="voy-detail" data-voy-detail="${esc(id)}"><header><b>${esc(point.type)} · 敌人 Lv.${point.enemyLv}</b><span>历史最佳 ${starsHTML(best, 13)}</span></header><p>${esc(point.desc)}</p><div class="voy-meta"><b>体力 ${point.cost}</b>${left !== null ? `<b>今日余 ${left}/${point.daily_limit} 次</b>` : ""}<span>${point.waves} 波</span></div><h4>敌人阵容</h4>${mobs}${mechanic ? `<p>${esc(mechanic)}</p>` : ""}<h4>三星目标</h4><p>战斗胜利 · 全员存活 · ${point.starRound} 回合内取胜</p><h4>胜利奖励</h4>${rewards}${milestones}${sweep}<p class="voy-gate" role="status">${gate ? `解锁条件：${esc(gate)}` : ""}</p><div class="modal-btns"><button class="gbtn gbtn-sm" type="button" data-voy-formation><span>布阵</span></button><button class="gbtn gbtn-primary" type="button" data-voy-start="${esc(id)}"><span>挑战 · ${point.cost} 体力</span></button></div></article>`);
  const detail = document.querySelector("[data-voy-detail]");
  detail.dataset.voyDay = date;
  detail.querySelector("[data-voy-formation]").onclick = () => { closeModal(); nav("page-squad"); };
  detail.querySelector("[data-voy-start]").onclick = () => { closeModal(); startBattle(id); };
  detail.querySelectorAll("[data-voy-sweep]").forEach(button => { button.onclick = () => voyageSweep(detail, id, Number(button.dataset.voySweep)); });
  updateStageRunLimits();
  document.querySelector("#modal-box .mb-close")?.focus();
}

function voyageUpdateRunButton(seconds) {
  const detail = document.querySelector("[data-voy-detail]"), button = detail?.querySelector("[data-voy-start]");
  if (!button || !Store.ready()) return;
  const point = voyageStage(button.dataset.voyStart), gate = voyageGate(point);
  const reason = gate ? `解锁条件：${gate}` : voyageRemaining(point) === 0 ? "今日次数已用完，北京时间零点刷新" : seconds ? `连续挑战休整中，${seconds} 秒后可出战` : SAVE.stamina < point.cost ? "体力不足，可回港领取时段补给" : "";
  button.disabled = !!reason;
  detail.querySelector(".voy-gate").textContent = reason;
  const sweepStatus = detail.querySelector("[data-voy-sweep-status]");
  if (sweepStatus) {
    const left = voyageRemaining(point), deny = detail.dataset.voyDay !== SAVE.daily.date ? "已进入新的游戏日，请重新打开航点" : (SAVE.stars[point.id] || 0) < 3 ? "本航点取得三星后可扫荡" : Store.activeRun ? "请先结算进行中的战斗" : reason;
    detail.querySelectorAll("[data-voy-sweep]").forEach(control => {
      const count = Number(control.dataset.voySweep);
      control.disabled = !!deny || Store.mutationBusy || SAVE.stamina < point.cost * count || left !== null && left < count;
    });
    sweepStatus.textContent = deny || "";
  }
}

async function voyageSweep(detail, id, count) {
  const button = detail.querySelector(`[data-voy-sweep="${count}"]`);
  if (!button || button.disabled || !requireGame()) return;
  const epoch = Store.sessionEpoch;
  detail.querySelectorAll("[data-voy-sweep]").forEach(control => { control.disabled = true; });
  const receipt = await Store.mutate("voyage_sweep", {item_id:id, count, choice:detail.dataset.voyDay});
  if (epoch !== Store.sessionEpoch || !Store.ready()) return;
  if (!receipt) { if (detail.isConnected) updateStageRunLimits(); return; }
  renderTopbar();
  if (document.body.dataset.page === "page-stage") {
    renderStage();
    if (detail.isConnected && !document.getElementById("modal-mask").classList.contains("hidden")) voyageDetail(id);
  }
  toast((receipt.drops || []).join(" · ") || "扫荡已结算");
}

function voyageChest(id) {
  const context = voyageContext(), chest = context?.chests.find(candidate => candidate.id === id);
  if (!chest) return;
  const stars = voyageTotal(context.ids), claimed = !!SAVE.voyage_claims?.[id];
  showModal(`<h3>${esc(context.title)} · ${chest.stars} 星箱</h3><article class="voy-detail"><p>${esc(VOYAGE_MODES[context.view.mode])} · 本章已收集 ${stars} / ${context.ids.length*3} 星</p><div class="voy-drop-grid">${voyagePrizes(chest.prize)}</div><p>各航点只计历史最高星。章节、难度分别记账，每箱限领一次。</p><div class="modal-btns"><button class="gbtn gbtn-primary" type="button" data-voy-claim ${claimed || stars < chest.stars ? "disabled" : ""}><span>${claimed ? "已领取" : stars < chest.stars ? `还差 ${chest.stars - stars} 星` : "领取奖励"}</span></button></div></article>`);
  const button = document.querySelector("[data-voy-claim]");
  button.onclick = async () => {
    if (button.disabled || !Store.ready()) return;
    const epoch = Store.sessionEpoch;
    button.disabled = true;
    const result = await Store.mutate("voyage_chest", {item_id:id});
    if (epoch !== Store.sessionEpoch || !Store.ready()) return;
    if (result) {
      if (button.isConnected) closeModal();
      if (document.body.dataset.page === "page-stage") renderStage();
      renderTopbar(); toast((result.drops || []).join(" · ") || "航路星箱已领取");
    } else if (button.isConnected) button.disabled = false;
  };
  document.querySelector("#modal-box .mb-close")?.focus();
}

function voyageResultSummary(id, result) {
  const book = Store.profile?.voyage;
  if (!book || !result.win || !result.save?.stars) return "";
  for (const chapter of book.chapters) {
    for (const [mode, ids] of Object.entries(chapter.routes)) {
      if (!ids.includes(id)) continue;
      const total = voyageTotal(ids, result.save), ready = chapter.chests[mode].filter(chest => total >= chest.stars && !result.save.voyage_claims?.[chest.id]);
      return `<p class="voy-result">${esc(chapter.name)} · ${esc(VOYAGE_MODES[mode])} ${total}/${ids.length*3} ★${ready.length ? ` · ${ready.length} 档星箱可领` : ""}</p>`;
    }
  }
  return "";
}

window.matchMedia?.("(max-width: 650px) and (orientation: portrait)")?.addEventListener?.("change", () => {
  if (document.body.dataset.page === "page-stage" && Store.ready()) renderStage();
});

document.addEventListener("keydown", event => {
  if (event.key !== "Escape" || document.body.dataset.page !== "page-stage" || document.getElementById("modal-mask")?.classList.contains("hidden")) return;
  const point = document.querySelector("[data-voy-detail]")?.dataset.voyDetail;
  if (!point && !document.querySelector("[data-voy-claim]")) return;
  closeModal();
  if (point) Array.from(document.querySelectorAll("[data-voy-point]")).find(button => button.dataset.voyPoint === point)?.focus();
});
