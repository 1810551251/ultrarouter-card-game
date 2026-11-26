let captainBoundaryTimer = null;

const captainDestinations = {voyage:["page-stage","去出航"],drill:["page-growth","去训练"],wish:["page-growth","去许愿"],outfit:["page-growth","去培养"]};

function captainJournal() {
  const captain = SAVE?.captain, rules = Store.profile?.captain_rules, log = SAVE?.daily?.captain_log;
  if (!captain || !rules || !log) return null;
  const clock = new Date(Store.serverNow() + 8 * 3600000), day = SAVE.daily.date;
  return { captain, rules, log, day, clock, hour:clock.getUTCHours(), current:day === clock.toISOString().slice(0,10) };
}

function captainSupplyState(journal, supply) {
  if (!journal.current) return "新日待同步";
  if (journal.log.supplies[supply.id]) return "已领取";
  if (journal.hour < supply.start) return "未开餐";
  return journal.hour < supply.end ? "领取体力" : "已收餐";
}

function captainTasksReady(journal) {
  return journal.current ? journal.rules.tasks.filter(task => !journal.log.claimed[task.id] && (journal.log.progress[task.id] || 0) >= task.target).length : 0;
}

function captainWatchBoundary(journal) {
  clearTimeout(captainBoundaryTimer);
  if (!journal?.current || !["page-captain","page-home"].includes(document.body.dataset.page)) return;
  const next = Math.min(24,...journal.rules.supplies.flatMap(slot => [slot.start,slot.end]).filter(hour => hour > journal.hour));
  const clock = journal.clock, delay = ((next-journal.hour)*60-clock.getUTCMinutes())*60000-clock.getUTCSeconds()*1000-clock.getUTCMilliseconds()+50;
  const epoch = Store.sessionEpoch;
  captainBoundaryTimer = setTimeout(() => {
    if (epoch !== Store.sessionEpoch || !Store.ready()) return;
    if (document.body.dataset.page === "page-captain") renderCaptain();
    if (["page-home","page-captain"].includes(document.body.dataset.page)) renderCaptainSummary();
  },delay);
}

function renderCaptainSummary() {
  const display = document.getElementById("res-captain"), summary = document.getElementById("captain-summary");
  const journal = captainJournal();
  if (display) display.textContent = journal ? `Lv.${journal.captain.level}` : "日志";
  if (summary) {
    if (!journal) summary.textContent = "今日航务待同步";
    else {
      const supply = journal.rules.supplies.find(slot => captainSupplyState(journal,slot) === "领取体力");
      const ready = captainTasksReady(journal);
      summary.textContent = !journal.current ? "新航日已到 · 同步日志" : `${ready ? `${ready} 项任务可领奖` : "今日任务已登记"} · ${supply ? `${supply.name}可领` : "查看三餐补给"}`;
    }
  }
  captainWatchBoundary(journal);
}

function renderCaptain() {
  const body = document.getElementById("captain-body");
  if (!body || !Store.ready()) return;
  const journal = captainJournal();
  const syncButton = '<button type="button" class="gbtn gbtn-sm" data-captain-sync><span>同步日志</span></button>';
  body.onclick = captainJournalClick;
  body.dataset.captainDay = journal?.day || "";
  body.dataset.captainEpoch = String(Store.sessionEpoch);
  if (!journal) { body.innerHTML = `<p>船长日志尚未同步。</p>${syncButton}`; return; }
  const {captain,rules,log,day,current} = journal, full = captain.level >= GROWTH.maxLevel;
  const cost = full ? 1 : rules.xp_base + rules.xp_step*captain.level;
  const claimed = rules.tasks.filter(task => log.claimed[task.id]).length;
  body.innerHTML = `<section class="captain-passport"><div><p class="social-kicker">航籍 · ${esc(day)}</p><h3>船长 Lv.${captain.level}</h3><p>伙伴等级上限 <b>Lv.${captain.level}</b></p></div><div class="captain-xp"><label>船长经验 <strong>${full ? "满级" : `${captain.xp} / ${cost}`}</strong><progress value="${full ? 1 : captain.xp}" max="${cost}"></progress></label><small>副本胜利、经验训练和每日任务可积累航行经验。</small></div></section>
    <section class="captain-supplies" aria-labelledby="captain-galley"><header class="mode-card-head"><div><p class="social-kicker">北京时间 · 满体力亦可领取</p><h3 id="captain-galley">航海厨房</h3></div>${syncButton}</header>
    ${!current ? '<p role="status">已进入新的航海日，同步后可查看今日航务。</p>' : ""}<div class="captain-meals">${rules.supplies.map(slot => {
      const status = captainSupplyState(journal,slot), open = status === "领取体力";
      return `<article class="captain-meal" data-supply-state="${open ? "ready" : "closed"}"><small>${String(slot.start).padStart(2,"0")}:00–${String(slot.end).padStart(2,"0")}:00</small><h4>${esc(slot.name)}</h4><p>体力 +${slot.stamina}</p><button type="button" class="gbtn gbtn-sm${open ? " gbtn-primary" : ""}" data-captain-supply="${esc(slot.id)}" ${open ? "" : "disabled"}><span>${status}</span></button></article>`;
    }).join("")}</div></section>
    <section class="captain-duties" aria-labelledby="captain-orders"><header class="mode-card-head"><div><p class="social-kicker">每日 00:00 换日 · 已领 ${claimed}/${rules.tasks.length}</p><h3 id="captain-orders">今日航务</h3></div></header>${rules.tasks.map(task => {
      const progress = log.progress[task.id] || 0, taken = log.claimed[task.id], ready = current && !taken && progress >= task.target;
      const destination = !ready && !taken && current && captainDestinations[task.id];
      return `<article class="captain-duty"><div><h4>${esc(task.name)} <small>${progress}/${task.target}</small></h4><p>${esc(task.goal)}</p><small>船长经验 +${task.xp} · 金币 +${task.gold}</small></div><button type="button" class="gbtn gbtn-sm${ready ? " gbtn-primary" : ""}" ${destination ? `data-captain-visit="${esc(task.id)}"` : `data-captain-task="${esc(task.id)}"`} ${ready || destination ? "" : "disabled"}><span>${taken ? "已领取" : ready ? "领取奖励" : destination ? destination[1] : current ? "进行中" : "待同步"}</span></button></article>`;
    }).join("")}</section>`;
  captainWatchBoundary(journal);
}

async function captainJournalClick(event) {
  const button = event.target.closest("button");
  if (!button || button.disabled || !Store.ready()) return;
  const epoch = Store.sessionEpoch, journal = captainJournal();
  const body = button.closest("#captain-body");
  if (!body || body.dataset.captainEpoch !== String(epoch)) return;
  if (button.dataset.captainVisit && captainDestinations[button.dataset.captainVisit]) { nav(captainDestinations[button.dataset.captainVisit][0]); return; }
  let receipt = null;
  button.disabled = true;
  try {
    if (button.hasAttribute("data-captain-sync")) await Store.refresh();
    else if (journal?.current && body.dataset.captainDay === journal.day) {
      const taskID = button.dataset.captainTask, supplyID = button.dataset.captainSupply;
      const task = journal.rules.tasks.find(rule => rule.id === taskID), supply = journal.rules.supplies.find(rule => rule.id === supplyID);
      if (task && !journal.log.claimed[taskID] && journal.log.progress[taskID] >= task.target) {
        receipt = await Store.mutate("captain_task",{item_id:`${journal.day}:${taskID}`});
      } else if (supply && captainSupplyState(journal,supply) === "领取体力") {
        receipt = await Store.mutate("captain_supply",{item_id:`${journal.day}:${supplyID}`});
      }
    }
  } finally {
    if (epoch === Store.sessionEpoch && Store.ready()) {
      if (document.body.dataset.page === "page-captain") renderCaptain();
      renderTopbar();
      if (receipt) toast((receipt.drops || []).join(" · ") || "已领取");
    }
  }
}
