const fleetTabs = { crew: "我的公会", directory: "招募名录" };
const fleetScreen = { tab: "directory", epoch: -1, request: 0, pages: [0], next: 0 };

function fleetNumber(value) {
  return (Number.isSafeInteger(value) && value >= 0 ? value : 0).toLocaleString("zh-CN");
}

function fleetButton(label, command, value = "", disabled = false) {
  return `<button type="button" class="fleet-button" data-fleet-act="${command}" data-value="${esc(value)}" ${disabled ? "disabled" : ""}>${esc(label)}</button>`;
}

function fleetEnrollment() {
  return `<div class="fleet-enroll"><details><summary>持邀请码入会</summary><form data-fleet-form="join"><label>邀请码<input name="code" placeholder="8 位邀请码" minlength="8" maxlength="8" pattern="[a-fA-F0-9]{8}" autocapitalize="characters" required></label><button class="fleet-button" type="submit">加入公会</button></form></details>
    <details><summary>创建公会</summary><form data-fleet-form="create"><label>公会名称<input name="name" minlength="2" maxlength="12" required></label><label class="fleet-check"><input name="listed" type="checkbox" checked> 公开招募并参加公会榜</label><button class="fleet-button" type="submit">成立公会</button></form></details></div>`;
}

function fleetCrew(guild) {
  if (!guild) return `<div class="fleet-empty"><h3>尚未入会</h3>${fleetButton("前往招募名录", "tab", "directory")}</div>${fleetEnrollment()}`;
  const harbor = guild.harbor, leader = Store.profile.user_id === guild.leader_id;
  const members = Array.isArray(guild.members) ? guild.members : [];
  const bossHP = Math.max(0, guild.boss_hp || 0), bossMax = Math.max(1, guild.boss_max_hp || 1);
  const bossReady = !guild.attacked_today && bossHP > 0 && SAVE.stamina >= 10 && SAVE.squad.length > 0;
  const bossLabel = bossHP <= 0 ? "首领已击败" : guild.attacked_today ? "今日已出战" : SAVE.stamina < 10 ? "体力不足" : !SAVE.squad.length ? "尚未编队" : "出战 · 10 体力";
  return `<div class="fleet-banner"><div><small>公会旗号 · ${esc(guild.boss_week)}</small><h3>${esc(guild.name)}</h3></div><div><small>本周航绩</small><b>${fleetNumber(harbor?.score)}</b></div><div><small>船员</small><b>${members.length} / 10</b></div></div>
    ${harbor ? `<section class="fleet-section"><div class="fleet-section-title"><h3>母港共建</h3><small>本周已交付 ${fleetNumber(harbor.my_cargo)} 次 · ${harbor.supplied ? "今日已交付" : "今日可选一项"}</small></div><div class="fleet-lanes">${harbor.lines.map((line, index) => `<article><small>工位 0${index + 1}</small><h4>${esc(line.name)}</h4><progress aria-label="${esc(line.name)}进度" value="${line.done}" max="${harbor.line_limit}"></progress><small>${fleetNumber(line.done)} / ${fleetNumber(harbor.line_limit)} 份</small>${fleetButton(`${line.cost} · 交付`, "supply", line.id, !line.eligible)}</article>`).join("")}</div>
    <h4 class="fleet-minor-title">共建里程</h4><p class="fleet-note">本周在本会交付过一次即可领取已达成的里程。三个工位须同时达标；每档每账号每周限领一次。</p><div class="fleet-list">${harbor.beacons.map(beacon => `<div class="fleet-row"><span class="fleet-ordinal">${beacon.tier}</span><div><b>${esc(beacon.name)}</b><small>各 ${beacon.goal} 份 · 金币 ${beacon.gold}、星锻矿 ${beacon.ore}</small></div>${fleetButton(beacon.claimed ? "已领取" : beacon.eligible ? "领取物资" : "未达成", "claim", String(beacon.tier), !beacon.eligible)}</div>`).join("")}</div></section>` : ""}
    <section class="fleet-section"><div class="fleet-section-title"><h3>协力讨伐</h3><small>${esc(ELEM[guild.boss_element]?.name || "元素")}系首领</small></div><p class="fleet-note">剩余生命 ${fleetNumber(bossHP)} / ${fleetNumber(bossMax)} · 我的伤害 ${fleetNumber(guild.my_damage)}</p><progress class="fleet-boss-hp" aria-label="首领剩余生命" value="${bossHP}" max="${bossMax}"></progress><div class="fleet-actions">${fleetButton(bossLabel, "attack", "", !bossReady)}${fleetButton(guild.reward_claimed ? "已领取讨伐奖励" : "领取讨伐奖励", "bossClaim", "", !guild.can_claim)}</div><p class="fleet-note">本周击败首领后，参战成员可领 250 金币和 3 枚元素结晶。</p></section>
    <details class="fleet-details"><summary>船员名册、邀请码与公会设置</summary><div class="fleet-roster">${members.map(member => `<span>${esc(member.name)}${member.user_id === guild.leader_id ? " · 会长" : ""}</span>`).join("")}</div><div class="fleet-actions">${fleetButton(`邀请码 ${guild.code} · 复制`, "copy", guild.code)}${leader ? fleetButton(guild.listed ? "停止公开招募" : "开启公开招募", "listing", guild.listed ? "off" : "on") : ""}</div><p class="fleet-note">${guild.listed ? "已公开招募并参加公会榜。" : "未列入公开名录，仍可凭邀请码加入。"}</p><details><summary>退出公会</summary><p class="fleet-note">${leader && members.length > 1 ? "仍有其他成员，会长暂不能退出。" : "退出会清除在本会的个人贡献，今日次数与本周已领奖励不会重置。"}</p>${fleetButton("确认退出", "leave", "", leader && members.length > 1)}</details></details>`;
}

function fleetBoardMarkup(view) {
  const member = Boolean(Store.profile?.guild), limit = view.member_limit;
  const list = view.entries.map(entry => {
    const canJoin = !member && entry.members < limit && /^[0-9A-F]{8}$/.test(entry.code || "");
    return `<li class="fleet-row${entry.mine ? " is-mine" : ""}"><span class="fleet-ordinal">帆</span><div><b>${esc(entry.name)}${entry.mine ? "<em>本会</em>" : ""}</b><small>${fleetNumber(entry.members)} / ${fleetNumber(limit)} 位成员</small></div><div class="fleet-value"><b>${fleetNumber(entry.value)}</b><small>航绩</small></div>${canJoin ? fleetButton("加入", "recruit", entry.code) : ""}</li>`;
  }).join("");
  return `<section class="fleet-section"><div class="fleet-section-title"><h3>招募名录</h3><small>第 ${fleetScreen.pages.length} 页</small></div><p class="fleet-note">公开招募 · 仅列未满员公会</p>
    ${list ? `<ol class="fleet-list">${list}</ol>` : '<p class="fleet-empty">本页暂无招募中的公会，可凭邀请码加入或自行建会。</p>'}
    <div class="fleet-pager">${fleetButton("‹ 上一页", "previous", "", fleetScreen.pages.length < 2)}${fleetButton("下一页 ›", "next", "", !view.next)}</div></section>${!member ? fleetEnrollment() : ""}`;
}

function fleetCurrent(request, epoch) {
  return request === fleetScreen.request && Store.sessionEpoch === epoch && Store.ready() && document.body.dataset.page === "page-fleet";
}

async function fleetLoad(syncProfile = false) {
  const pane = document.getElementById("fleet-pane"), tab = fleetScreen.tab;
  if (!pane) return;
  const epoch = Store.sessionEpoch, request = ++fleetScreen.request;
  document.querySelectorAll("#fleet-body [data-fleet-tab]").forEach(button => button.setAttribute("aria-pressed", String(button.dataset.fleetTab === tab)));
  pane.innerHTML = '<p class="fleet-empty" role="status">正在读取港务记录…</p>';
  try {
    if (syncProfile || tab === "crew") {
      if (!await Store.refresh()) throw new Error("暂未取得最新公会记录");
      if (!fleetCurrent(request, epoch)) return;
      renderTopbar();
    }
    if (tab === "crew") { pane.innerHTML = fleetCrew(Store.profile?.guild); return; }
    const after = tab === "directory" ? fleetScreen.pages.at(-1) : 0;
    const data = await Store.api(`${SITE.ep.fleet}?board=${tab}&after=${after}`, "GET");
    if (!fleetCurrent(request, epoch)) return;
    if (data?.board !== tab || !Array.isArray(data.entries)) throw new Error("榜单记录暂不可用");
    fleetScreen.next = data.next || 0;
    pane.innerHTML = fleetBoardMarkup(data);
  } catch (error) {
    if (fleetCurrent(request, epoch)) pane.innerHTML = `<div class="fleet-empty"><p>${esc(error.message)}</p>${fleetButton("重新读取", "refresh")}</div>`;
  }
}

async function fleetMutate(action, fields, message) {
  const epoch = Store.sessionEpoch, request = fleetScreen.request;
  const result = await Store.mutate(action, fields);
  if (!result || !fleetCurrent(request, epoch)) return;
  if (["guild_join", "guild_create", "guild_leave"].includes(action)) {
    fleetScreen.tab = action === "guild_leave" ? "directory" : "crew";
    fleetScreen.pages = [0];
  }
  renderTopbar();
  toast((result.drops || []).join(" · ") || message);
  await fleetLoad();
}

function fleetClick(event) {
  const button = event.target.closest("button");
  if (!button || button.disabled) return;
  const command = button.dataset.fleetAct, value = button.dataset.value || "";
  const tab = button.dataset.fleetTab || (command === "tab" && value);
  if (tab && Object.hasOwn(fleetTabs, tab)) {
    fleetScreen.tab = tab; fleetScreen.pages = [0]; return fleetLoad();
  }
  switch (command) {
    case "hall": return openFleetRanking("power", button);
    case "refresh": return fleetLoad(true);
    case "previous": if (fleetScreen.pages.length > 1) { fleetScreen.pages.pop(); return fleetLoad(); } break;
    case "next": if (fleetScreen.next) { fleetScreen.pages.push(fleetScreen.next); return fleetLoad(); } break;
    case "recruit": if (/^[0-9A-F]{8}$/.test(value)) return fleetMutate("guild_join", { item_id: value }, "已加入公会"); break;
    case "supply": return fleetMutate("guild_supply", { item_id: value }, "交付完成");
    case "claim": return fleetMutate("guild_supply_claim", { count: Number(value) }, "里程奖励已领取");
    case "listing": return fleetMutate("guild_listing", { enabled: value === "on" }, "招募状态已更新");
    case "attack": return fleetMutate("guild_attack", {}, "讨伐完成");
    case "bossClaim": return fleetMutate("guild_claim", {}, "讨伐奖励已领取");
    case "leave": return fleetMutate("guild_leave", {}, "已退出公会");
    case "copy": return copyGuildCode(value);
  }
}

function fleetSubmit(event) {
  const form = event.target;
  if (!form.dataset.fleetForm) return;
  event.preventDefault();
  const data = new FormData(form);
  if (form.dataset.fleetForm === "join") {
    const code = String(data.get("code") || "").trim().toUpperCase();
    if (!/^[0-9A-F]{8}$/.test(code)) return toast("请输入 8 位邀请码");
    return fleetMutate("guild_join", { item_id: code }, "已加入公会");
  }
  const name = String(data.get("name") || "").trim();
  if (Array.from(name).length < 2 || Array.from(name).length > 12) return toast("公会名称需为 2 至 12 个字");
  return fleetMutate("guild_create", { item_id: name, enabled: data.has("listed") }, "公会已成立");
}

function renderFleet() {
  if (!requireGame()) return;
  if (fleetScreen.epoch !== Store.sessionEpoch) {
    fleetScreen.epoch = Store.sessionEpoch;
    fleetScreen.tab = Store.profile?.guild ? "crew" : "directory";
    fleetScreen.pages = [0];
  }
  const root = document.getElementById("fleet-body");
  root.innerHTML = `<div class="fleet-nav"><div role="group" aria-label="港务厅栏目">${Object.entries(fleetTabs).map(([tab, name]) => `<button type="button" data-fleet-tab="${tab}">${name}</button>`).join("")}${fleetButton("航海风云榜", "hall")}</div><button type="button" class="fleet-reload" data-fleet-act="refresh" aria-label="刷新港务厅" title="刷新港务厅">↻</button></div><div id="fleet-pane"></div>`;
  root.onclick = fleetClick;
  root.onsubmit = fleetSubmit;
  return fleetLoad();
}

const fleetRankKinds = {
  power: ["战斗力", "POWER", "战力", "当前上阵编队的综合战力。"], collection: ["收藏率", "COLLECTION", "收藏率", "已获得的不同整卡占图鉴总数的比例，碎片不计。"],
  guild: ["公会排行", "UNION", "本周航绩", "参与公开招募的公会：每 100 点讨伐伤害计 1 航绩，每次共建计 2 航绩。"], arena: ["赏金竞技", "ARENA", "竞技胜场", "只计已结算的竞技胜场，平局不计。"],
  stars: ["副本星级榜", "ADVENTURE", "副本星数", "四章剧情、困难、噩梦逐关取历史最高星。同一关不重复计数；资源副本与裂隙不计。"],
};
const fleetRank = { mode: "", board: "power", request: 0, view: null, opener: null, detailRow: null, fromLive: false };

function fleetRankExample(board) {
  const cast = [
    ["霁湾来信", "薄荷灯塔", 91, ["c20", "c14", "c03", "c27"], 81670, 25, 1084, 1756, 147],
    ["风笺十七", "纸鸢测候社", 84, ["c12", "c04", "c18", "c24"], 79240, 26, 864, 1682, 132],
    ["银杯听潮", "绒鲸邮局", 79, ["c16", "c21", "c05", "c09"], 76190, 23, 956, 1540, 138],
    ["半枚潮印", "雾屿船票行", 72, ["c08", "c27", "c02", "c25"], 64810, 24, 695, 1492, 121],
    ["栗壳信标", "双弦渡口", 68, ["c17", "c06", "c22", "c13"], 59760, 20, 738, 1324, 105],
    ["蓝檐回声", "砾石巡潮队", 63, ["c09", "c11", "c19", "c01"], 57480, 21, 633, 1266, 95],
    ["墨斗晚钟", "弯月帆铺", 59, ["c22", "c10", "c05", "c14"], 53620, 17, 528, 1130, 86],
    ["曙砂折页", "霜灯图书船", 56, ["c06", "c02", "c16", "c07"], 48950, 19, 436, 1064, 72],
  ];
  const metric = {power:4, collection:5, guild:6, arena:7, stars:8}[board];
  const entries = cast.map((row, index) => ({ name: row[board === "guild" ? 1 : 0], level: row[2], value: row[metric], mine: index === 5, members: 10 - index % 4,
    team: row[3].map((id, slot) => ({ id, level: row[2] - slot * 2, star: 4 + (index + slot) % 2 })) })).sort((left, right) => right.value - left.value);
  const mine = entries.findIndex(entry => entry.mine);
  return { board, entries, catalog_size: 28, star_limit: 156, member_limit: 10, my_name: entries[mine].name, my_level: entries[mine].level, my_value: entries[mine].value, my_rank: mine + 1 };
}

function fleetRankScore(board, value, view) {
  if (board === "collection") return view.catalog_size ? `${(100 * value / view.catalog_size).toFixed(1)}%` : "0.0%";
  return `${fleetNumber(value)}${board === "stars" ? " ★" : board === "arena" ? " 胜" : ""}`;
}

function fleetRankCards(entry) {
  if (!Array.isArray(entry.team) || typeof GIRLS === "undefined") return [];
  return entry.team.slice(0, 4).flatMap(card => {
    const girl = GIRLS.find(girl => girl.id === card?.id);
    return girl ? [{ girl, level: card.level, star: card.star }] : [];
  });
}

function fleetRankFaces(cards) {
  return cards.length ? `<span class="rank-faces">${cards.map(({girl, level}) => `<span><img src="${esc(girl.img)}" alt="${esc(girl.name)}" title="${esc(girl.name)}"><small>Lv.${fleetNumber(level)}</small></span>`).join("")}</span>` : '<span class="rank-unformed">未编队</span>';
}

function fleetRankRows(view) {
  const board = view.board, guild = board === "guild", kind = fleetRankKinds[board];
  const items = view.entries.map((entry, index) => {
    const cards = fleetRankCards(entry), captain = cards[0]?.girl;
    const caption = guild ? `${fleetNumber(entry.members)} / ${fleetNumber(view.member_limit)} 位成员` : entry.level ? `Lv.${fleetNumber(entry.level)}` : "船长";
    const extra = board === "collection" ? `${fleetNumber(entry.value)} / ${fleetNumber(view.catalog_size)} 张整卡` : board === "stars" ? `上限 ${fleetNumber(view.star_limit)} ★` : "";
    return `<li><button type="button" class="rank-row${entry.mine ? " rank-mine" : ""}" data-rank-row="${index}" aria-haspopup="dialog"><span class="rank-medal" data-medal="${index < 3 ? index + 1 : 0}">${index + 1}</span><span class="rank-player">${captain ? `<img class="rank-avatar" src="${esc(captain.img)}" alt="">` : '<span class="rank-avatar rank-emblem" aria-hidden="true">⚓</span>'}<span><b>${esc(entry.name)}</b><small>${esc(caption)}${entry.mine ? (guild ? " · 本会" : " · 我") : ""}</small></span></span><span class="rank-score">${fleetRankScore(board, entry.value, view)}${extra ? `<small>${extra}</small>` : ""}</span>${fleetRankFaces(cards)}</button></li>`;
  }).join("");
  const standing = view.my_rank ? `第 ${fleetNumber(view.my_rank)} 名` : "暂未上榜";
  return `<div class="rank-columns"><span>排名</span><span>${guild ? "公会" : "船长"}</span><span>${kind[2]}</span><span>${guild ? "会长编队" : "队伍"}</span></div><div class="rank-scroll" tabindex="0" aria-label="排行榜名单">${items ? `<ol class="fleet-list rank-list">${items}</ol>` : `<div class="rank-empty"><b>暂无上榜记录</b><p>${kind[3]}</p><button type="button" data-rank-action="demo">查看演示榜</button></div>`}</div><footer class="fleet-standing rank-standing"><span>${guild ? "本会名次" : "我的名次"}</span><b>${standing}</b><span>${esc(view.my_name || "")}${view.my_level ? ` · ${guild ? "会长 " : ""}Lv.${fleetNumber(view.my_level)}` : ""}</span><strong>${kind[2]}：${fleetRankScore(board, view.my_value || 0, view)}</strong></footer>`;
}

function fleetOpenDialog(dialog) {
  if (typeof dialog.showModal === "function") dialog.showModal(); else dialog.setAttribute("open", "");
}

function fleetRemoveDialog(dialog) {
  if (typeof dialog?.close === "function" && dialog.open) dialog.close();
  dialog?.remove();
}

function closeFleetRankDetail() {
  const detail = document.getElementById("fleet-rank-detail");
  if (!detail) return;
  fleetRemoveDialog(detail);
  fleetRank.detailRow?.focus(); fleetRank.detailRow = null;
}

function closeFleetRanking() {
  const hall = document.getElementById("fleet-rank-hall");
  if (!hall) return;
  closeFleetRankDetail(); fleetRank.request++; fleetRank.view = null; fleetRank.mode = "";
  fleetRemoveDialog(hall);
  if (/^#rank-demo(?:-|$)/.test(location.hash)) history.replaceState(history.state, "", location.href.split("#")[0]);
  const fallback = document.querySelector('[data-rank-demo-launch], #title-screen:not(.hidden) [onclick^="openFleetRankDemo"], #app:not(.hidden) .page.active button, #title-screen:not(.hidden) #auth-actions:not(.hidden) button');
  (fleetRank.opener?.isConnected ? fleetRank.opener : fallback)?.focus();
  fleetRank.opener = null;
}

function closeFleetOfficial() {
  if (fleetRank.mode === "live") closeFleetRanking();
  else if (fleetRank.fromLive) {
    fleetRank.fromLive = false;
    const source = document.querySelector('#fleet-rank-hall [data-rank-action="source"]');
    if (source) source.hidden = true;
  }
}

function showFleetRankDetail(row, button) {
  const view = fleetRank.view, entry = view?.entries[row];
  if (!entry || !button || document.getElementById("fleet-rank-detail")) return;
  const dialog = document.createElement("dialog"), kind = fleetRankKinds[view.board], cards = fleetRankCards(entry);
  dialog.id = "fleet-rank-detail"; dialog.className = "rank-detail"; dialog.setAttribute("aria-labelledby", "rank-detail-name");
  dialog.innerHTML = `<header><small>${fleetRank.mode === "demo" ? "演示榜 · 虚构档案" : "正式榜 · 公开档案"}</small><h3 id="rank-detail-name">${esc(entry.name)}</h3></header><p>第 ${fleetNumber(row + 1)} 名 · ${kind[2]} ${fleetRankScore(view.board, entry.value, view)}</p><h4>${view.board === "guild" ? "会长编队" : "上阵队伍"}</h4>${fleetRankFaces(cards)}<p class="rank-rule">${kind[3]}</p><button type="button" data-rank-action="close-detail">返回排行榜</button>`;
  fleetRank.detailRow = button;
  document.getElementById("fleet-rank-hall").append(dialog);
  dialog.addEventListener("cancel", event => { event.preventDefault(); closeFleetRankDetail(); });
  fleetOpenDialog(dialog); dialog.querySelector("button").focus();
}

async function fleetRankLoad(board) {
  if (!Object.hasOwn(fleetRankKinds, board)) return;
  const hall = document.getElementById("fleet-rank-hall");
  if (!hall) return;
  closeFleetRankDetail();
  const demo = fleetRank.mode === "demo", request = ++fleetRank.request, epoch = demo ? null : Store.sessionEpoch;
  const current = () => hall.isConnected && request === fleetRank.request && (demo || (Store.ready() && epoch === Store.sessionEpoch));
  fleetRank.board = board; fleetRank.view = null;
  hall.dataset.rankBoard = board; hall.dataset.rankSource = demo ? "demo" : "live";
  hall.querySelectorAll("[data-fleet-tab]").forEach(button => button.setAttribute("aria-pressed", String(button.dataset.fleetTab === board)));
  hall.querySelector("[data-rank-origin]").textContent = demo ? "演示榜 · 虚构船长与公会" : "正式榜 · 全服";
  const toggle = hall.querySelector('[data-rank-action="source"]'); toggle.hidden = demo && !fleetRank.fromLive; toggle.textContent = demo ? "返回正式榜" : "查看演示榜";
  const pane = hall.querySelector(".rank-content");
  pane.innerHTML = '<p class="rank-empty" role="status">正在读取榜单…</p>';
  try {
    const view = demo ? fleetRankExample(board) : await Store.api(`${SITE.ep.fleet}?board=${board}&after=0`, "GET");
    if (!current()) return;
    if (!view || view.board !== board || !Array.isArray(view.entries)) throw new Error("暂未取得排行榜");
    fleetRank.view = view; pane.innerHTML = fleetRankRows(view);
  } catch (error) {
    if (current()) pane.innerHTML = `<div class="rank-empty" role="status"><b>暂未取得正式榜</b><p>${esc(error.message)}</p><button type="button" data-rank-action="retry">重新读取</button><button type="button" data-rank-action="demo">查看演示榜</button></div>`;
  }
}

function fleetRankClick(event) {
  const button = event.target.closest("button");
  if (!button) return;
  if (button.dataset.fleetTab) return fleetRankLoad(button.dataset.fleetTab);
  if (button.hasAttribute("data-rank-row")) return showFleetRankDetail(Number(button.dataset.rankRow), button);
  switch (button.dataset.rankAction) {
    case "back": return closeFleetRanking();
    case "close-detail": return closeFleetRankDetail();
    case "retry": return fleetRankLoad(fleetRank.board);
    case "demo": case "source": {
      if (fleetRank.mode === "live") { fleetRank.fromLive = true; fleetRank.mode = "demo"; }
      else if (fleetRank.fromLive && Store.ready()) fleetRank.mode = "live";
      else return;
      document.querySelector(`#fleet-rank-hall [data-fleet-tab="${fleetRank.board}"]`).focus();
      return fleetRankLoad(fleetRank.board);
    }
  }
}

function openFleetHall(board, opener, demo) {
  if (!Object.hasOwn(fleetRankKinds, board)) return;
  let hall = document.getElementById("fleet-rank-hall");
  if (!hall) {
    hall = document.createElement("dialog"); hall.id = "fleet-rank-hall"; hall.setAttribute("aria-labelledby", "rank-hall-title");
    hall.innerHTML = `<header class="rank-heading"><button class="rank-back" type="button" data-rank-action="back" aria-label="返回上一页" title="返回上一页">↶</button><h2 id="rank-hall-title">航海风云榜<small>HALL OF FAME</small></h2><div class="rank-source"><b data-rank-origin></b><button type="button" data-rank-action="source"></button></div></header><div class="rank-layout"><nav class="rank-sidebar" aria-label="排行榜分类">${Object.entries(fleetRankKinds).map(([key, kind]) => `<button type="button" data-fleet-tab="${key}"><b>${kind[0]}</b><small>${kind[1]}</small></button>`).join("")}</nav><section class="rank-content" aria-label="当前排行榜"></section></div>`;
    hall.addEventListener("click", fleetRankClick);
    hall.addEventListener("cancel", event => { if (event.target === hall) { event.preventDefault(); closeFleetRanking(); } });
    hall.addEventListener("keydown", event => {
      const active = event.target.dataset.fleetTab;
      if (!active || !["ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
      event.preventDefault();
      const tabs = Object.keys(fleetRankKinds), index = tabs.indexOf(active), direction = ["ArrowUp", "ArrowLeft"].includes(event.key) ? -1 : 1;
      const next = tabs[event.key === "Home" ? 0 : event.key === "End" ? tabs.length - 1 : (index + direction + tabs.length) % tabs.length];
      hall.querySelector(`[data-fleet-tab="${next}"]`).focus(); fleetRankLoad(next);
    });
    fleetRank.opener = opener?.isConnected ? opener : document.activeElement !== document.body ? document.activeElement : null;
    document.body.append(hall); fleetOpenDialog(hall);
  }
  fleetRank.fromLive = false; fleetRank.mode = demo ? "demo" : "live";
  hall.querySelector(`[data-fleet-tab="${board}"]`).focus();
  return fleetRankLoad(board);
}

function openFleetRanking(board = "power", opener) { return openFleetHall(board, opener, !Store.ready()); }
function openFleetRankDemo(board = "stars", opener) { return openFleetHall(board, opener, true); }

function fleetRankHash() {
  const match = /^#rank-demo(?:-(power|collection|guild|arena|stars))?$/.exec(location.hash);
  if (match) openFleetRankDemo(match[1] || "stars");
}
window.addEventListener("hashchange", fleetRankHash);
if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", fleetRankHash, {once:true}); else fleetRankHash();
