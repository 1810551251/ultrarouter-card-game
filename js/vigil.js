const VIGIL_SIGNS = {lantern:"✧",tide:"≋",ward:"⟐"};
let vigilPending = null;

function vigilEquipment(need) {
  const stock = SAVE.casebook?.evidence || {}, names = Store.profile?.cases?.evidence || {};
  return Object.entries(need || {}).map(([id,cost]) => ({id,name:names[id] || id,cost,have:stock[id] || 0}));
}

function vigilQuote(id, choice, aspectID) {
  const rules = Store.profile?.vigil, own = SAVE?.girls[id];
  if (!rules || !own?.star) return null;
  const card = own.vigil, rank = card?.rank || 0, marks = card?.marks || 0, strain = card?.strain || 0;
  const route = rules.paths.find(path => path.id === (card?.path || choice)) || rules.paths[0];
  const selected = route.aspects.find(aspect => aspect.id === (aspectID || card?.aspect));
  const current = route.aspects.find(aspect => aspect.id === card?.aspect), next = rules.ranks[rank];
  const sealStock = SAVE.mats.break_seal || 0, ascent = vigilEquipment(rank === 1 ? selected?.evidence : next?.evidence);
  const convert = vigilEquipment(selected?.evidence), grounding = vigilEquipment({[rules.ground_evidence]:1});
  const paid = rows => rows.every(row => row.have >= row.cost);
  const groups = new Map(), eligible = SAVE.squad.includes(id) && rank > 0;
  for (const member of SAVE.squad) {
    const oath = SAVE.girls[member]?.vigil;
    if (!oath || oath.rank < 1) continue;
    if (!groups.has(oath.path)) groups.set(oath.path,new Set());
    if (oath.aspect) groups.get(oath.path).add(oath.aspect);
  }
  const twins = groups.get(route.id)?.size || 0;
  return {rules,card,rank,marks,strain,route,selected,current,next,sealStock,ascent,convert,grounding,twins,groups,
    trineActive:eligible && groups.size === rules.paths.length,twinsActive:eligible && twins >= 2,
    daily:SAVE.daily.vigil_marks?.[id] || 0,base:rules.ranks[rank-1]?.effect || 0,
    canBind:!card,canAscend:!!card && !!next && marks >= next.marks && sealStock >= next.seals && (rank !== 1 || !!selected) && paid(ascent),
    canRefocus:rank >= 2 && !!selected && selected.id !== card.aspect && sealStock >= rules.refocus_seals && paid(convert),
    canGround:strain > 0 && paid(grounding),canOmen:rank > 0 && !card.overdraw && strain <= rules.strain_start_max,canRest:rank > 0 && card.overdraw};
}

function selectVigilPath(id) {
  const view = currentWorkshop();
  if (view?.tab !== "vigil" || SAVE.girls[view.id].vigil || !Store.profile?.vigil?.paths.some(path => path.id === id)) return;
  view.vigilChoice = id; view.vigilAspect = ""; paintWorkshop(`[data-vigil-path="${id}"]`);
}

function selectVigilAspect(id) {
  const view = currentWorkshop();
  if (view?.tab !== "vigil" || !vigilQuote(view.id,view.vigilChoice)?.route.aspects.some(aspect => aspect.id === id)) return;
  view.vigilAspect = id; paintWorkshop(`[data-vigil-aspect="${id}"]`);
}

async function vigilAction(action, enabled) {
  const view = currentWorkshop();
  if (view?.tab !== "vigil" || vigilPending?.epoch === view.epoch) return;
  const quote = vigilQuote(view.id,view.vigilChoice,view.vigilAspect);
  if (!quote || (action === "vigil_stance" ? typeof enabled !== "boolean" : enabled !== undefined)) return;
  const allowed = {vigil_bind:quote.canBind,vigil_ascend:quote.canAscend,vigil_refocus:quote.canRefocus,vigil_ground:quote.canGround,vigil_stance:enabled ? quote.canOmen : quote.canRest};
  if (!Object.hasOwn(allowed,action) || !allowed[action]) return;
  const fields = {girl_id:view.id};
  if (action === "vigil_bind") fields.choice = quote.route.id;
  if (action === "vigil_refocus" || action === "vigil_ascend" && quote.rank === 1) fields.choice = quote.selected.id;
  if (action === "vigil_stance") fields.enabled = enabled;
  const token = {epoch:view.epoch};
  vigilPending = token;
  let result;
  try { paintWorkshop(); result = await Store.mutate(action,fields); }
  finally {
    if (vigilPending === token) vigilPending = null;
    if (Store.ready() && Store.sessionEpoch === token.epoch) {
      if (result) refreshCultivationViews(view.id);
      if (currentWorkshop()?.tab === "vigil" && document.querySelector('#modal-mask:not(.hidden) [data-vigil-panel]') && (!result || currentWorkshop().id !== view.id)) paintWorkshop();
    }
  }
  if (result && Store.ready() && Store.sessionEpoch === token.epoch) toast((result.drops || []).join(" · ") || "仪轨已确认");
}

function vigilInventory(rows) {
  return rows.map(row => `<span class="vg-supply" data-vigil-evidence="${esc(row.id)}" data-ready="${row.have >= row.cost}">${esc(row.name)} <b>${row.have}/${row.cost}</b></span>`).join("");
}

function vigilWorkshopPanel(id, choice) {
  const view = currentWorkshop(), quote = vigilQuote(id,choice,view?.vigilAspect);
  if (!quote) return "";
  const {rules,route,card,selected,current,rank,next,marks,strain,base,sealStock} = quote;
  const busy = vigilPending?.epoch === Store.sessionEpoch;
  const command = (action,label,ok) => `<button type="button" class="ws-command" data-vigil-action="${action}" onclick="vigilAction('${action}')" ${ok && !busy ? "" : "disabled"}>${esc(label)}</button>`;
  const seals = cost => `<span class="vg-supply" data-ready="${sealStock >= cost}">晋阶章 <b>${sealStock}/${cost}</b></span>`;
  const ranks = [{name:"立誓"},...rules.ranks];
  const paths = rules.paths.map(path => `<button type="button" data-vigil-path="${path.id}" aria-pressed="${path.id === route.id}" ${card ? "disabled" : ""} onclick="selectVigilPath('${path.id}')"><i aria-hidden="true">${VIGIL_SIGNS[path.id]}</i><b>${esc(path.name)}</b></button>`).join("");
  const stance = (over,isCurrent) => {
    const stanceValue = rank ? base+(over ? rules.omen_bonus : 0) : 0;
    return `<section data-vigil-mode="${over ? "overdraw" : "normal"}" data-active="${isCurrent}"><header><b>${over ? "借夜" : "守常"}</b><span>${isCurrent ? "当前" : "备选"}</span></header><strong>+${stanceValue}</strong><span>${esc(route.effect_label)}</span>${current ? `<small>${esc(current.effect_label)} +${stanceValue}</small>` : ""}<small>${over ? `入战付最大生命 ${rules.omen_hp_cost}%（不致命）；胜利失序 +${rules.strain_gain}` : "入战保留生命；胜利失序 −1"}</small><button type="button" class="ws-command" data-vigil-stance="${over ? "overdraw" : "normal"}" aria-pressed="${isCurrent}" onclick="vigilAction('vigil_stance',${over})" ${!busy && (over ? quote.canOmen : quote.canRest) ? "" : "disabled"}>${isCurrent ? "已启用" : over ? "开启借夜" : "收起借夜"}</button></section>`;
  };
  const selection = route.aspects.map(aspect => `<button type="button" data-vigil-aspect="${aspect.id}" aria-pressed="${selected?.id === aspect.id}" onclick="selectVigilAspect('${aspect.id}')"><span>${current?.id === aspect.id ? "当前分支" : "分支"} · ${esc(aspect.name)}</span><b>${esc(aspect.effect_label)}</b><small>${esc(aspect.story)}</small><span class="vg-supplies">${vigilInventory(vigilEquipment(aspect.evidence))}</span></button>`).join("");
  return `<section class="vg-sheet" data-vigil-panel="${id}" data-route="${route.id}"><header class="vg-heading"><div><small class="ws-kicker">夜巡仪轨</small><h4>${esc(card ? route.name : "选择誓约")}</h4></div><strong data-vigil-rank>${card ? `${rank} 阶 · ${esc(ranks[rank].name)}` : "未立誓"}</strong></header><div class="vg-routes" aria-label="誓约路线">${paths}</div><p class="vg-motto">${esc(route.motto)}</p><div class="vg-task"><b>见证</b><span>${esc(route.task)}</span></div>${!card ? `<div class="vg-bind">${command("vigil_bind",`立誓 · ${route.name}`,true)}<small>免费 · 路线不可更改</small></div>` : ""}
  <section class="vg-section"><h4>仪式进度 <small data-vigil-marks>累计 ${marks} 印</small></h4><ol class="vg-ranks">${ranks.map((entry,index) => `<li data-complete="${!!card && rank >= index}"><i></i><b>${esc(entry.name)}</b><small>${index ? `${entry.marks} 印` : "起点"}</small></li>`).join("")}</ol><div class="vg-task"><b data-vigil-daily>今日 ${quote.daily}/${rules.daily_limit}</b><span>${SAVE.squad.includes(id) ? "本卡已入队" : "本卡未入队"} · 仅随队完成见证才获印，晋阶不扣印。</span></div>${next ? `<p class="vg-need">下一阶：${esc(next.name)} · 累计 ${marks}/${next.marks} 印 ${seals(next.seals)} ${vigilInventory(quote.ascent)}</p>${rank === 1 && !selected ? "<p>晋第二阶须先选定分支。</p>" : ""}${command("vigil_ascend","完成下一阶仪式",quote.canAscend)}` : "<p>三阶仪式已完成。</p>"}</section>
  <section class="vg-section"><h4>分支定相 <small>${esc(current?.name || "第二阶选择")}</small></h4><div class="vg-aspects">${selection}</div>${rank >= 2 ? `<p>改换同路分支：${seals(rules.refocus_seals)} ${selected && selected !== current ? vigilInventory(quote.convert) : "另付目标分支证物"}</p>${command("vigil_refocus","改换所选分支",quote.canRefocus)}` : ""}<button type="button" class="ws-link" data-vigil-cases onclick="selectWorkshopTab('cases')">调查异闻 · 查找证物 ›</button></section>
  <section class="vg-section"><h4>夜巡姿态 <small>${rank ? "对下一场战斗生效" : "第一阶解锁"}</small></h4><div class="vg-stances">${stance(false,!!card && !card.overdraw)}${stance(true,!!card?.overdraw)}</div><div class="vg-strain"><label>失序 <b data-vigil-strain>${strain}/${rules.strain_max}</b> <progress max="${rules.strain_max}" value="${strain}"></progress></label><small>${rules.strain_max} 强制熄灯，降至 ${rules.strain_start_max} 或以下可再借夜。</small></div><div class="vg-bind">${command("vigil_ground",`镇定 −${rules.ground_drop} 失序`,quote.canGround)} ${vigilInventory(quote.grounding)}</div></section>
  <section class="vg-section"><h4>上阵共鸣</h4><div class="vg-formations"><div data-vigil-formation="trine" data-active="${quote.trineActive}"><b>三誓齐阵 · ${quote.groups.size}/${rules.paths.length}</b><span>三路一阶以上同队，各誓约者开场 +${rules.chorus_rage} 怒气。</span></div><div data-vigil-formation="twins" data-active="${quote.twinsActive}"><b>同路异相 · ${quote.twins}/2</b><span>同路两种分支上阵，该路誓约者开场获 ${rules.twins_shield}% 生命护盾。</span></div></div></section></section>`;
}
