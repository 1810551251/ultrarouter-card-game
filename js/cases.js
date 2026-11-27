const CASE_CHOICES = {survey:"取证",veil:"潜行",dare:"冒险"};
const CASE_SIGNS = {ember:"✧",echo:"≋",key:"⌘"};

function caseAmounts(rules,values) {
  return Object.entries(rules.evidence).map(([key,title])=>`<span data-case-evidence="${key}">${typeof atelierIcon === "function" && atelierItem(key) ? atelierIcon(key) : `<i aria-hidden="true">${CASE_SIGNS[key]}</i>`}${esc(title)} <b>${values?.[key] || 0}</b></span>`).join("");
}

function caseOutcomeText(rules, report) {
  if (!report || !rules.files[report.case_id]) return "";
  return `<div class="case-outcome" data-case-ending="${esc(report.ending)}"><b>${esc(rules.endings[report.ending])}</b> · ${esc(rules.files[report.case_id].name)}<small>${Object.entries(report.awards || {}).length ? Object.entries(report.awards).map(([id,count])=>`${esc(rules.evidence[id])} ×${count}`).join(" · ") : "无证物入库"}</small></div>`;
}

function casePressureMark(delta) { return delta>0 ? `+${delta}` : String(delta); }

function caseActPanel(rules,book) {
  const active=book.active, file=rules.files[active.case_id], step=active.trail.length, act=file.acts[step];
  const leader=GIRLS.find(girl=>girl.id===active.leader);
  const oath=rules.oaths[active.oath];
  const choices=act ? `<div class="case-moves">${act.options.map(option=>{
    const relief=!!(option.relief && active.clues[option.relief]), tide=!active.oath_used && active.oath==="tide" && option.id==="dare", lantern=!active.oath_used && active.oath==="lantern" && option.id==="survey";
    const delta=option.pressure-Number(relief)-Number(tide), predicted=Math.max(0,active.pressure+delta);
    const guarded=predicted>=rules.pressure_limit && active.oath==="ward" && !active.oath_used;
    return `<button type="button" data-case-step="${option.id}" onclick="caseWorkshopAction('step','${option.id}')"><small>${CASE_CHOICES[option.id]}</small><b>${esc(option.name)}</b><span>${esc(option.text)}</span><em>${esc(rules.evidence[option.clue])} +${option.amount+Number(lantern)} · 压力 ${casePressureMark(delta)}${guarded ? " · 触发守阈" : predicted>=rules.pressure_limit ? " · 将失控" : ""}</em>${option.relief ? `<small>${esc(rules.evidence[option.relief])} ${relief ? "已互证，压力 −1" : "可互证减压"}</small>` : ""}</button>`;
  }).join("")}</div>` : "";
  const trail=active.trail.map((choice,index)=>`${index+1}. ${esc(file.acts[index].options.find(option=>option.id===choice)?.name || choice)}`).join(" › ");
  return `<section class="case-active"><header><small class="ws-kicker">${esc(file.name)} · ${act ? `第 ${step+1} 幕 / ${file.acts.length}` : "三幕已勘查"}</small><h4>${act ? esc(act.title) : "结案抉择"}</h4><p>${act ? esc(act.text) : "现场已封存。余下证物能否还原完整经过，取决于已相互印证的线索。"}</p></header><div class="case-gauges"><label>压力 <b>${active.pressure}/${rules.pressure_limit}</b> <meter min="0" max="${rules.pressure_limit}" value="${active.pressure}"></meter></label><span>领队：${esc(leader?.name || "独立调查")}</span></div>${oath ? `<p class="case-oath" data-case-oath="${active.oath}">${esc(oath)} · <b>${active.oath_used ? "已使用" : "待触发"}</b></p>` : ""}<div class="case-ledger">${caseAmounts(rules,active.clues)}</div>${trail ? `<small class="case-trail">${trail}</small>` : ""}${choices}<div class="case-close"><button type="button" class="ws-command" data-case-close="file" onclick="caseWorkshopAction('close','file')" ${step ? "" : "disabled"}>保守归档 · 至多 ${rules.file_limit} 份</button><button type="button" class="ws-command" data-case-close="pursue" onclick="caseWorkshopAction('close','pursue')" ${step===3 ? "" : "disabled"}>追查真相 · 压力 +2</button><button type="button" class="ws-link" data-case-close="withdraw" onclick="caseWorkshopAction('close','withdraw')">撤离 · 不保留证物</button></div><small>追查：${esc(rules.evidence[file.focus])} 至少 2 份、至少两种线索、合计 5 份。真相最多留 ${rules.pursuit_limit} 份；残页最多 ${rules.partial_limit} 份；压力达 ${rules.pressure_limit} 即失控。</small></section>`;
}

function casesWorkshopPanel(girl,own) {
  const rules=Store.profile?.cases, book=SAVE?.casebook;
  if (!rules || !book) return "";
  const leader=own?.vigil?.rank>0 ? rules.oaths[own.vigil.path] : "";
  const opened=book.opened || [], locked=opened.length>=rules.daily_limit;
  const files=rules.order.map((id,index)=>{
    const file=rules.files[id], visited=opened.includes(id), endings=book.endings?.[id] || {};
    return `<article><small>案 ${String(index+1).padStart(2,"0")} · ${esc(rules.evidence[file.focus])}</small><h4>${esc(file.name)}</h4><p>${esc(file.intro)}</p>${Object.keys(endings).length ? `<small>${Object.entries(endings).map(([ending,count])=>`${esc(rules.endings[ending])} ×${count}`).join(" · ")}</small>` : ""}<button type="button" class="ws-command" data-case-open="${id}" onclick="caseWorkshopAction('open','${id}')" ${locked || visited ? "disabled" : ""}>${visited ? "今日已立案" : locked ? "今日名额已满" : "调阅立案"}</button></article>`;
  }).join("");
  return `<section class="case-desk" data-cases-panel><header><small class="ws-kicker">夜巡值案 · ${esc(book.date)}</small><h4>异闻调查局</h4><p>今日立案 <b>${opened.length}/${rules.daily_limit}</b> · 证物入库 <b>${book.awarded || 0}/${rules.daily_budget}</b></p></header><div class="case-ledger">${caseAmounts(rules,book.evidence)}</div>${book.active ? caseActPanel(rules,book) : `<div class="case-leader"><label><input type="checkbox" data-case-leader checked> ${esc(girl.name)} 领队</label><small>${esc(leader || "无可用调查誓约")}</small></div><div class="case-files">${files}</div>`}${caseOutcomeText(rules,book.last)}</section>`;
}

async function caseWorkshopAction(kind,id) {
  const view=currentWorkshop(), rules=Store.profile?.cases, book=SAVE?.casebook;
  if (!view || view.tab!=="cases" || !rules || !book) return;
  const active=book.active, fields={};
  if (kind==="open") {
    if (active || !rules.order.includes(id) || book.opened?.includes(id) || (book.opened?.length || 0)>=rules.daily_limit) return;
    fields.item_id=id;
    if (document.querySelector("[data-case-leader]")?.checked) fields.girl_id=view.id;
  } else {
    if (!active) return;
    const step=active.trail.length;
    if (kind==="step" && rules.files[active.case_id].acts[step]?.options.some(option=>option.id===id)) fields.count=step+1;
    else if (kind==="close" && (id==="withdraw" || id==="file" && step>0 || id==="pursue" && step===3)) fields.count=step;
    else return;
    fields.item_id=active.id;fields.choice=id;
  }
  const epoch=Store.sessionEpoch, result=await Store.mutate(`case_${kind}`,fields);
  if (!result || epoch!==Store.sessionEpoch || !Store.ready()) return;
  refreshCultivationViews(view.id);
  toast((result.drops || []).join(" · ") || "调查档案已更新");
}
