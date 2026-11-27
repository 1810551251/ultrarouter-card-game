/* Ranked arena is an independent server-owned ladder; no E stake or local save. */
let ladderData = null;
let ladderSampledAt = 0;
let ladderRankPage = 0;
let ladderRecordFilter = "all";
let ladderQuote = null;

const ladderNow = () => ladderData ? ladderData.server_now + performance.now() - ladderSampledAt : Store.serverNow();
const ladderTime = value => new Date(value).toLocaleString("zh-CN", {timeZone:"Asia/Shanghai", hour12:false});
const ladderMaterials = materials => `<div class="ladder-materials">${["mat_r", "mat_sr", "mat_ssr"].map(key => `<span><i aria-hidden="true">◆</i>${esc(MATS[key]?.name || key)} <b>×${Number(materials?.[key]) || 0}</b></span>`).join("")}</div>`;

async function renderLadder() {
  const epoch = Store.sessionEpoch, version = ++arenaQueueVersion;
  const tab = arenaTab;
  const current = () => Store.ready() && epoch === Store.sessionEpoch && version === arenaQueueVersion && document.body.dataset.page === "page-arena" && arenaTab === tab;
  document.getElementById("arena-body").innerHTML = '<p class="social-empty" role="status">正在读取竞技排名与结算奖励…</p>';
  try {
    const profile = await Store.refresh();
    if (!current()) return;
    if (!profile) throw new Error("暂时无法同步游戏资料，请重试。");
    const data = await Store.api(SITE.ep.ladder, "GET");
    if (!current()) return;
    ladderData = data;
    ladderSampledAt = performance.now();
    paintLadderSummary(); renderTopbar(); paintLadder();
    tickLadderClock();
  } catch (e) {
    if (!current()) return;
    document.getElementById("arena-body").innerHTML = `<div class="social-empty"><p>${esc(e.message)}</p><button class="gbtn" onclick="renderArena()"><span>重新读取竞技场</span></button></div>`;
  }
}

function paintLadderSummary() {
  const me = ladderData?.me;
  const leader = GIRLS.find(g => g.id === (me?.squad?.[0]?.id || SAVE?.squad?.[0]));
  document.getElementById("arena-summary").innerHTML = `<div class="arena-captain">${leader ? `<img src="${leader.img}" alt="${leader.name}">` : '<span class="arena-emblem">✦</span>'}<div><span>星霜排名竞技场</span><strong>${esc(SITE_ACCOUNT?.name || "星霜旅人")}</strong><span>当前进攻战力</span><b>${(SAVE?.power || 0).toLocaleString()}</b></div></div><div class="arena-metrics"><div><span>当前排名</span><b>${me ? `#${me.rank}` : "未报名"}</b></div><div><span>历史最高</span><b>${me ? `#${me.best_rank}` : "—"}</b></div><div><span>今日挑战</span><b>${ladderData?.remaining ?? "—"}<small> / ${ladderData?.limit || 5}</small></b></div><div><span>每日结算</span><b>21:00</b></div></div><div class="arena-sidebar-squad"><span>已保存的防守阵容</span><div class="arena-squad">${arenaSquadHTML(me?.squad) || '<p>报名时保存当前编队</p>'}</div></div>`;
}

function ladderHeading(title, note) {
  return `<div class="arena-section-heading"><div><p class="social-kicker">RANKED ARENA · 每日 21:00</p><h3>${title}</h3><p>${note}</p></div><div><button class="social-refresh" onclick="renderArena()">刷新</button>${ladderData.me && arenaTab === "ladder" ? '<button class="social-refresh" onclick="confirmLadderLeave()">退出排名竞技</button>' : ""}</div></div>`;
}
function paintLadder() {
  if (!ladderData || !Store.ready()) return;
  if (arenaTab === "ladder-board") return paintLadderBoard();
  if (arenaTab === "ladder-rewards") return paintLadderRewards();
  if (arenaTab === "ladder-records") return paintLadderRecords();
  const {me, remaining, opponents} = ladderData;
  const tier = ladderData.rewards.find(t => !t.max_rank || me && me.rank <= t.max_rank);
  document.getElementById("arena-body").innerHTML = ladderHeading("向更高的名次出发", "挑战守卫与其他玩家的防守阵容 · 获胜晋位 · 不消耗 E 余额") +
    `<div class="ladder-settlement"><div><span>距离下次材料结算</span><b data-ladder-time>—</b><small>北京时间每日 21:00</small></div><div><span>${me ? `当前第 ${me.rank} 名 · 预计奖励` : "报名后参与每日排名奖励"}</span>${me ? ladderMaterials(tier?.materials) : '<p>微光结晶 · 元素结晶 · 星屑</p>'}</div></div>` +
    (!me ? `<div class="ladder-welcome"><span class="arena-empty-emblem">♜</span><h3>加入星霜竞技场</h3><p>从榜尾开始，逐步挑战更高排名的对手。报名会保存当前编队作为防守阵容，其他玩家可以挑战你的防守。</p><p>每日 5 次免费挑战，05:00 重置；21:00 按结算时名次发放材料。</p><button class="gbtn gbtn-primary" ${SAVE.squad.length ? "" : "disabled"} onclick="ladderAction('ladder_join')"><span>${SAVE.squad.length ? "免费报名 · 开启挑战" : "请先配置出战编队"}</span></button></div>` :
    `<div class="arena-list-head"><p>今日剩余 <b>${remaining} / ${ladderData.limit}</b> 次 · <span data-ladder-cooldown>准备就绪</span></p><button class="social-refresh" onclick="renderArena()">换一批对手</button></div><div class="ladder-opponents">${opponents.map(entry => ladderOpponentCard(entry)).join("") || '<p class="social-empty">暂无可挑战对手，稍后刷新看看。</p>'}</div><div class="ladder-defense-tools"><div><b>我的防守阵容 · 战力 ${me.power.toLocaleString()}</b><div class="arena-squad">${arenaSquadHTML(me.squad)}</div><p>养成或调整编队后，点击保存才会更新防守。进攻使用当前编队。</p></div><button class="gbtn" onclick="confirmLadderDefense()"><span>保存当前编队为防守</span></button></div>`);
}

function ladderOpponentCard(entry) {
  const girl = GIRLS.find(g => g.id === entry.squad?.[0]?.id);
  return `<article class="arena-opponent-card ladder-opponent"><div class="arena-opponent-top"><span class="ladder-rank">#${entry.rank}</span><span class="arena-room-badge">${entry.id < 0 ? "星霜守卫" : "玩家防守"}</span></div><button class="ladder-portrait" onclick="previewLadderOpponent(${entry.id})" aria-label="查看${esc(entry.name)}的阵容">${girl ? `<img src="${girl.img}" alt="${girl.name}" loading="lazy">` : "✦"}</button><h4>${esc(entry.name)}</h4><p>防守战力 <b>${entry.power.toLocaleString()}</b></p><div class="arena-squad">${arenaSquadHTML(entry.squad)}</div><button class="gbtn gbtn-primary" data-ladder-challenge ${ladderData.remaining <= 0 || ladderNow() < ladderData.me.cooldown_until || !SAVE.squad.length ? "disabled" : ""} onclick="previewLadderOpponent(${entry.id},true)"><span>发起挑战</span></button><small>${entry.rank < ladderData.me.rank ? `获胜可升至第 ${entry.rank} 名` : "获胜保持当前名次"}</small></article>`;
}
function previewLadderOpponent(id, challenge = false) {
  if (!requireGame() || !ladderData) return;
  const entry = [...ladderData.opponents, ...ladderData.board].find(x => x.id === id);
  if (!entry) return;
  ladderQuote = {id, rank:entry.rank, epoch:Store.sessionEpoch};
  showModal(`<h3>${challenge ? "确认排名挑战" : "竞技场防守阵容"}</h3><div class="social-confirm"><h4>#${entry.rank} · ${esc(entry.name)}</h4><div class="arena-comparison"><section><p>对方防守 · ${entry.power.toLocaleString()}</p><div class="arena-squad arena-preview">${arenaSquadHTML(entry.squad)}</div></section><section><p>当前进攻 · ${(SAVE.power || 0).toLocaleString()}</p><div class="arena-squad arena-preview">${arenaSquadHTML(arenaOwnSquad())}</div></section></div><p>获胜且对手排名更高时交换名次。挑战消耗 1 次机会，不消耗 E 余额；60 回合未击败对手视为进攻失败。</p></div><div class="modal-btns"><button class="gbtn" onclick="closeModal()"><span>返回</span></button>${challenge ? '<button class="gbtn gbtn-primary" onclick="submitLadderChallenge()"><span>确认挑战 · 1 次</span></button>' : ""}</div>`);
}
async function submitLadderChallenge() {
  if (!ladderQuote || ladderQuote.epoch !== Store.sessionEpoch) return;
  await ladderAction("ladder_challenge", {opponent_id:ladderQuote.id, count:ladderQuote.rank});
}
function confirmLadderDefense() {
  showModal(`<h3>更新防守阵容</h3><div class="social-confirm"><p>将当前编队的等级、星级和装备保存为防守快照，其他玩家挑战时使用这套阵容。</p><div class="arena-squad arena-preview">${arenaSquadHTML(arenaOwnSquad())}</div></div><div class="modal-btns"><button class="gbtn" onclick="closeModal()"><span>返回</span></button><button class="gbtn gbtn-primary" onclick="ladderAction('ladder_defense')"><span>确认保存</span></button></div>`);
}
async function ladderAction(action, fields = {}) {
  if (!requireGame()) return;
  const epoch = Store.sessionEpoch;
  const result = await Store.mutate(action,fields);
  if (!result || epoch !== Store.sessionEpoch || !Store.ready()) return;
  closeModal();
  if (document.body.dataset.page === "page-arena" && arenaTab.startsWith("ladder")) await renderArena();
  if (epoch !== Store.sessionEpoch || !Store.ready()) return;
  renderTopbar();
  if (result.ladder) showArenaResult(result.ladder);
  else if (action === "ladder_claim") toast("排名材料已领取至背包");
  else if (action === "ladder_defense") toast("防守阵容已保存");
  else if (action === "ladder_leave") toast("已退出排名竞技");
}

function confirmLadderLeave() {
  showModal('<h3>退出排名竞技？</h3><p>退榜后不再接受排名挑战，也不再获得新的排名奖励。已结算材料仍可领取，重回榜尾不会恢复当天次数。</p><div class="modal-btns"><button class="gbtn" onclick="closeModal()"><span>返回</span></button><button class="gbtn gbtn-primary" onclick="ladderAction(\'ladder_leave\')"><span>确认退榜</span></button></div>');
}

function paintLadderBoard() {
  const board = ladderData.board, pages = Math.max(1,Math.ceil(board.length/10));
  ladderRankPage = Math.max(0,Math.min(pages-1,ladderRankPage));
  document.getElementById("arena-body").innerHTML = ladderHeading("星霜竞技排名", `共 ${ladderData.total} 位参赛者 · 显示前 50 名 · 按挑战胜负晋位`) +
    `<div class="arena-podium">${board.slice(0,3).map((entry,i) => { const girl=GIRLS.find(g=>g.id===entry.squad?.[0]?.id); return `<button class="arena-podium-card podium-${i+1}" onclick="previewLadderOpponent(${entry.id})"><span class="arena-podium-place">0${i+1}</span><span class="arena-podium-title">${["星霜之巅","耀星之席","逐光之席"][i]}</span>${girl ? `<img src="${girl.img}" alt="${girl.name}">` : ""}<strong>${esc(entry.name)}</strong><span>防守战力 <b>${entry.power.toLocaleString()}</b></span></button>`; }).join("")}</div>
    <p class="social-note">我的排名 ${ladderData.me ? `#${ladderData.me.rank} · 历史最高 #${ladderData.me.best_rank}` : "尚未报名"}</p><div class="arena-list">${board.slice(ladderRankPage*10,(ladderRankPage+1)*10).map(entry=>`<article class="arena-row ${entry.id===Store.authUserId ? "is-self" : ""}"><span class="arena-rank">${entry.rank}</span><div class="arena-player"><strong>${esc(entry.name)}${entry.id===Store.authUserId ? " · 我" : ""}</strong><p>${entry.id<0 ? "星霜守卫" : "玩家"} · 防守战力 ${entry.power.toLocaleString()}</p></div><button class="gbtn gbtn-sm" onclick="previewLadderOpponent(${entry.id})"><span>查看阵容</span></button></article>`).join("")}</div><div class="arena-pagination"><button class="social-refresh" ${ladderRankPage===0?"disabled":""} onclick="changeLadderPage(-1)">上一页</button><span>${ladderRankPage+1} / ${pages}</span><button class="social-refresh" ${ladderRankPage+1>=pages?"disabled":""} onclick="changeLadderPage(1)">下一页</button></div>`;
}
function changeLadderPage(delta) { ladderRankPage += delta; paintLadderBoard(); }
function paintLadderRewards() {
  const {me, rewards, mail} = ladderData;
  document.getElementById("arena-body").innerHTML = ladderHeading("每日排名 · 材料补给", "北京时间 21:00 按当时排名结算，领取后进入材料背包") + `<div class="ladder-settlement"><div><span>下次结算倒计时</span><b data-ladder-time>—</b><small>${esc(ladderTime(ladderData.next_settlement))}（北京时间）</small></div><p>离线期间的奖励也会保留；每个结算日只能领取一次。</p></div><h3 class="ladder-subtitle">排名奖励档位</h3><div class="ladder-reward-tiers">${rewards.map((tier,i)=>{
    const min=i ? rewards[i-1].max_rank+1 : 1;
    const active=me && me.rank>=min && (!tier.max_rank || me.rank<=tier.max_rank);
    return `<article class="ladder-reward-tier ${active?"is-current":""}"><div><b>${tier.max_rank===min ? `第 ${min} 名` : !tier.max_rank ? `${min} 名及以后` : `${min}～${tier.max_rank} 名`}</b>${active?'<span>当前档位</span>':""}</div>${ladderMaterials(tier.materials)}</article>`;
  }).join("")}</div><h3 class="ladder-subtitle">结算记录与领取</h3><div class="arena-list">${mail.map(m=>`<article class="ladder-mail"><div><b>${esc(m.day)} · 第 ${m.rank} 名</b><p>21:00 排名结算</p>${ladderMaterials(m.materials)}</div><button class="gbtn gbtn-sm" ${m.claimed?"disabled":""} onclick="claimLadderMail('${esc(m.day)}')"><span>${m.claimed?"已领取":"领取材料"}</span></button></article>`).join("") || '<div class="social-empty">暂无结算记录。报名后，在下一次 21:00 结算时获得排名材料。</div>'}</div><p class="social-note">优先显示未领取奖励，最多显示 60 条；领取后刷新可继续查看更早的记录。材料达到背包上限时，奖励保留待领取。</p>`;
}
async function claimLadderMail(day) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(day)) return;
  await ladderAction("ladder_claim", {item_id:day});
}
function paintLadderRecords() {
  const history = ladderData.history.filter(m => ladderRecordFilter === "all" || (ladderRecordFilter === "attack" ? m.attacker_id===Store.authUserId : m.defender_id===Store.authUserId));
  document.getElementById("arena-body").innerHTML = ladderHeading("每次交锋 · 每次晋位", "最近 20 场攻防记录 · 回放不消耗次数、不重复发奖") + `<div class="arena-filters">${[["all","全部"],["attack","我的进攻"],["defense","我的防守"]].map(([key,label])=>`<button class="social-refresh" aria-pressed="${ladderRecordFilter===key}" onclick="filterLadderRecords('${key}')">${label}</button>`).join("")}</div><div class="arena-list">${history.map(m=>{
    const attack=m.attacker_id===Store.authUserId, win=attack ? m.attack_win : !m.attack_win;
    const before=attack ? m.attacker_before : m.defender_before, after=attack ? m.attacker_after : m.defender_after;
    const valid=/^[0-9a-f-]{36}$/i.test(m.id);
    return `<article class="ladder-record"><div class="arena-list-head"><b class="${win?"arena-win":""}">${attack?"进攻":"防守"}${win?"胜利":"失败"}</b><span>${esc(ladderTime(m.played_at))}</span></div><div class="ladder-record-versus"><span><small>进攻方</small><b>${esc(m.attacker_name)}</b></span><span aria-hidden="true">→</span><span><small>防守方</small><b>${esc(m.defender_name)}</b></span></div><div class="arena-list-head"><p>我的排名 #${before} → #${after}</p><button class="gbtn gbtn-sm" ${valid?"":"disabled"} onclick="replayLadder('${valid?m.id:""}')"><span>战斗回放</span></button></div></article>`;
  }).join("") || '<div class="social-empty">暂无此类战报，发起一次排名挑战吧。</div>'}</div>`;
}
function filterLadderRecords(filter) { if (!["all","attack","defense"].includes(filter)) return; ladderRecordFilter=filter; paintLadderRecords(); }
async function replayLadder(id) {
  if (!requireGame()) return;
  const epoch=Store.sessionEpoch;
  try {
    const result=await Store.api(SITE.ep.ladderReplay+"?request_id="+encodeURIComponent(id),"GET");
    if (epoch===Store.sessionEpoch && Store.ready()) showArenaResult(result);
  } catch(e) { if (epoch===Store.sessionEpoch && Store.ready()) toast(e.message); }
}
function showLadderResult(result) {
  showModal(`<h3>排名竞技 · ${result.win?"胜利":"失败"}</h3><div class="social-confirm"><h4>${esc(result.name)}</h4><div class="arena-award">#${result.own_before} → #${result.own_after}<span>我的竞技排名</span></div><p>${result.report.timed_out?"达到60回合上限，防守方获胜。":"本场排名已结算。"} 材料在每日21:00按届时排名发放。</p><p>回放左侧为进攻方「${esc(result.attacker_name)}」，右侧为防守方「${esc(result.defender_name)}」。</p></div><div class="modal-btns"><button class="gbtn" onclick="leaveArenaReplay()"><span>返回排名竞技</span></button><button class="gbtn gbtn-primary" onclick="watchArenaResult()"><span>观看战报</span></button></div>`);
}
function tickLadderClock() {
  clearTimeout(arenaPollTimer);
  if (!ladderData || !Store.ready() || !arenaTab.startsWith("ladder") || document.body.dataset.page!=="page-arena" || document.hidden) return;
  const now=ladderNow();
  if (now>=ladderData.next_settlement || now>=ladderData.next_reset) { renderArena(); return; }
  const seconds=Math.max(0,Math.ceil((ladderData.next_settlement-now)/1000));
  document.querySelectorAll("[data-ladder-time]").forEach(el=>el.textContent=[Math.floor(seconds/3600),Math.floor(seconds/60)%60,seconds%60].map(v=>String(v).padStart(2,"0")).join(":"));
  const cooldown=Math.max(0,Math.ceil(((ladderData.me?.cooldown_until || 0)-now)/1000));
  document.querySelectorAll("[data-ladder-cooldown]").forEach(el=>el.textContent=cooldown?`冷却 ${cooldown} 秒`:"准备就绪");
  document.querySelectorAll("[data-ladder-challenge]").forEach(el=>el.disabled=ladderData.remaining<=0 || cooldown>0 || !SAVE.squad.length);
  arenaPollTimer=setTimeout(tickLadderClock,1000);
}
function showLadderRules() {
  showModal(`<h3>排名竞技场规则</h3><div class="arena-rules"><section><h4>01 · 排名挑战</h4><p>免费报名后进入榜尾，系统推荐附近名次的对手。挑战对手保存的防守阵容，进攻获胜且对手名次更高时交换排名；失败或60回合未击败对手，排名不变。没有战力差限制，不消耗E余额。</p></section><section><h4>02 · 次数与防守</h4><p>每天5次主动挑战，北京时间05:00重置；每次挑战后冷却10秒。被其他玩家挑战不消耗次数。进攻使用当前编队，防守使用报名或最近一次保存的阵容快照。</p></section><section><h4>03 · 21点材料结算</h4><p>北京时间每天21:00按当时名次发放微光结晶、元素结晶和星屑，前往材料奖励栏目领取。离线也保留奖励，每个结算日只领取一次；领取或回放都不会再次改变排名。</p></section><section><h4>04 · 战报</h4><p>保留最近20场进攻与防守记录。回放使用当时的完整战斗记录，左侧进攻、右侧防守。余额匹配和自定义房间属于独立玩法，不影响此处排名和挑战次数。</p></section></div><div class="modal-btns"><button class="gbtn" onclick="closeModal()"><span>知道了</span></button></div>`);
}
