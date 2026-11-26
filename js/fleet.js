const fleetTabs = { crew: "我的公会", directory: "招募名录", guild: "公会排行", power: "战力排行", collection: "收藏率排行" };
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
  const board = view.board, entries = view.entries || [], limit = view.member_limit;
  const member = Boolean(Store.profile?.guild), directory = board === "directory";
  const format = value => board === "collection" ? `${view.catalog_size ? (100 * value / view.catalog_size).toFixed(1) : 0}%` : fleetNumber(value);
  const rules = { directory: "公开招募 · 仅列未满员公会", guild: "本周公开公会 · 每 100 点讨伐伤害计 1 航绩，每次共建计 2 航绩", power: "当前编队的实战属性战力", collection: "已获得的不同整卡 / 本期图鉴总数，碎片不计入" };
  const list = entries.map((entry, index) => {
    const guild = directory || board === "guild";
    const extra = guild ? `${fleetNumber(entry.members)} / ${fleetNumber(limit)} 位成员` : board === "collection" ? `${fleetNumber(entry.value)} / ${fleetNumber(view.catalog_size)} 张整卡` : "编队战力";
    const canJoin = !member && entry.members < limit && /^[0-9A-F]{8}$/.test(entry.code || "");
    return `<li class="fleet-row${entry.mine ? " is-mine" : ""}"><span class="fleet-ordinal">${directory ? "帆" : String(index + 1).padStart(2, "0")}</span><div><b>${esc(entry.name)}${entry.mine ? `<em>${guild ? "本会" : "我"}</em>` : ""}</b><small>${extra}</small></div><div class="fleet-value"><b>${format(entry.value)}</b>${guild ? "<small>航绩</small>" : ""}</div>${directory && canJoin ? fleetButton("加入", "recruit", entry.code) : ""}</li>`;
  }).join("");
  return `<section class="fleet-section"><div class="fleet-section-title"><h3>${fleetTabs[board]}</h3><small>${directory ? `第 ${fleetScreen.pages.length} 页` : "前 50 名"}</small></div><p class="fleet-note">${rules[board]}${directory ? "" : " · 同分按固定登记顺序排列"}</p>
    ${directory ? "" : `<div class="fleet-standing"><span>${board === "guild" ? "本会" : "我的"}名次</span><strong>${view.my_rank ? `第 ${fleetNumber(view.my_rank)} 名` : "暂未上榜"}</strong>${view.my_rank ? `<small>${format(view.my_value)}</small>` : ""}</div>`}
    ${list ? `<ol class="fleet-list">${list}</ol>` : `<p class="fleet-empty">${directory ? "本页暂无招募中的公会，可凭邀请码加入或自行建会。" : board === "guild" ? "本周尚无公开公会留下航绩。" : "暂时没有符合条件的航海员。"}</p>`}
    ${directory ? `<div class="fleet-pager">${fleetButton("‹ 上一页", "previous", "", fleetScreen.pages.length < 2)}${fleetButton("下一页 ›", "next", "", !view.next)}</div>` : ""}</section>${directory && !member ? fleetEnrollment() : ""}`;
}

function fleetCurrent(request, epoch) {
  return request === fleetScreen.request && Store.sessionEpoch === epoch && Store.ready() && document.body.dataset.page === "page-fleet";
}

async function fleetLoad(syncProfile = false) {
  const pane = document.getElementById("fleet-pane"), tab = fleetScreen.tab;
  if (!pane) return;
  const epoch = Store.sessionEpoch, request = ++fleetScreen.request;
  document.querySelectorAll("[data-fleet-tab]").forEach(button => button.setAttribute("aria-pressed", String(button.dataset.fleetTab === tab)));
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
  root.innerHTML = `<div class="fleet-nav"><div role="group" aria-label="港务厅栏目">${Object.entries(fleetTabs).map(([tab, name]) => `<button type="button" data-fleet-tab="${tab}">${name}</button>`).join("")}</div><button type="button" class="fleet-reload" data-fleet-act="refresh" aria-label="刷新港务厅" title="刷新港务厅">↻</button></div><div id="fleet-pane"></div>`;
  root.onclick = fleetClick;
  root.onsubmit = fleetSubmit;
  return fleetLoad();
}
