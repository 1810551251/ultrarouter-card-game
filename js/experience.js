function upgradeRules() { return Store.profile?.training?.experience; }

function upgradeEstimate(own, selected) {
  const cap = cardLevelLimit(), spent = {};
  let room = -(own.exp || 0), experience = 0;
  for (let level = own.level; level < cap; level++) room += cardLevelCost(level);
  for (const material of upgradeRules().materials) {
    const count = Math.min(selected[material.id] || 0,Math.ceil(Math.max(0,room)/material.xp));
    if (count < 1) continue;
    spent[material.id] = count;
    const gain = Math.min(room,count*material.xp);
    experience += gain; room -= gain;
  }
  let level = own.level, exp = (own.exp || 0) + experience;
  while (level < cap && exp >= cardLevelCost(level)) { exp -= cardLevelCost(level); level++; }
  return { level, exp:level === cap ? 0 : exp, spent, experience };
}

function upgradeWorkshopSheet(view, girl, own) {
  const cap = cardLevelLimit(), full = own.level >= cap, cost = cardLevelCost(own.level);
  return `<section class="ws-paper xp-chapter" data-upgrade-sheet="${girl.id}"><header><small class="ws-kicker">成长手册 · 角色等级</small><h4>历练研读</h4><p>为${esc(girl.name)}选一批书卷，一次结算，可连续跨级。</p></header>
    <div class="xp-ledger"><small>当前等级 / 卡牌上限</small><strong>Lv.${own.level} <em>/ ${cap}</em></strong><progress aria-label="当前等级经验" value="${full ? 1 : own.exp || 0}" max="${full ? 1 : cost}"></progress><span>${full ? cap < GROWTH.maxLevel ? "提升船长等级可继续培养" : "等级已圆满" : `${(own.exp || 0).toLocaleString()} / ${cost.toLocaleString()} 经验`}</span></div>
    <button type="button" data-upgrade-open class="ws-command" onclick="openUpgradePicker()" ${full ? "disabled" : ""}>${full ? "已达等级上限" : "一键升级 · 选择材料"}</button><div class="xp-chapters">${[1,60,120].map((level,index) => `<div><b>${["启程","远征","星轨"][index]} · Lv.${level}</b><small>本级升一级 ${cardLevelCost(level).toLocaleString()} 经验</small></div>`).join("")}</div>
    <p>每实际消耗 1 点体力，获得 1 点船长经验。每日任务另奖经验。伙伴可培养至船长当前等级，最高 ${GROWTH.maxLevel} 级。</p></section>`;
}

function activeUpgradePicker() {
  const dialog = document.querySelector("dialog[data-upgrade-picker]"), view = currentWorkshop();
  return dialog && Store.ready() && view?.tab === "upgrade" && dialog.dataset.upgradePicker === view.id && dialog.dataset.epoch === String(Store.sessionEpoch) ? dialog : null;
}

function openUpgradePicker() {
  const view = currentWorkshop(), own = SAVE?.girls[view?.id], rules = upgradeRules();
  if (!Store.ready() || view?.tab !== "upgrade" || !own?.star || !rules || own.level >= cardLevelLimit() || activeUpgradePicker()) return;
  const pane = document.querySelector(`[data-upgrade-sheet="${view.id}"]`);
  if (!pane) return;
  const dialog = document.createElement("dialog");
  dialog.dataset.upgradePicker = view.id; dialog.dataset.epoch = String(Store.sessionEpoch);
  dialog.className = "xp-picker"; dialog.setAttribute("aria-labelledby","xp-picker-title");
  dialog.innerHTML = `<header><div><small>历练行囊 · 经验材料</small><h4 id="xp-picker-title">为${esc(GIRLS.find(girl => girl.id === view.id).name)}选择书卷</h4></div><button type="button" class="ws-command" data-feed-close aria-label="关闭材料选择">关闭</button></header>
    ${rules.materials.map(material => {
      const item = atelierItem(material.id), stock = SAVE.mats[material.id] || 0;
      return `<div class="xp-material" data-feed-row="${material.id}">${atelierIcon(material.id)}<label for="xp-${material.id}"><b>${esc(item.name)}</b><small>${material.xp.toLocaleString()} 经验 · 持有 ${stock}</small></label><input id="xp-${material.id}" type="number" min="0" max="${stock}" step="1" inputmode="numeric" value="${stock}" data-feed-material="${material.id}" ${stock ? '' : 'disabled'}><input type="checkbox" aria-label="使用${esc(item.name)}" data-feed-check="${material.id}" ${stock ? 'checked' : 'disabled'}></div>`;
    }).join("")}<button type="button" class="xp-select-all" data-feed-all>取消全选</button><output data-feed-preview aria-live="polite"></output><small>先用经验较少的书，满级后不再扣书。最后一本超出的经验不保留。</small><button type="button" class="ws-command" data-upgrade-submit disabled>确认一键升级</button>`;
  pane.append(dialog);
  dialog.addEventListener("close",() => { dialog.remove(); pane.querySelector("[data-upgrade-open]")?.focus({preventScroll:true}); },{once:true});
  dialog.querySelector("[data-feed-close]").onclick = () => { if (dialog.close) dialog.close(); else dialog.remove(); };
  const inputs = [...dialog.querySelectorAll("[data-feed-material]")];
  for (const input of inputs) {
    input.oninput = refreshUpgradePreview;
    const check = dialog.querySelector(`[data-feed-check="${input.dataset.feedMaterial}"]`);
    check.onchange = () => { input.value = check.checked ? input.max : '0'; refreshUpgradePreview(); };
  }
  dialog.querySelector("[data-feed-all]").onclick = () => {
    const clear = inputs.every(input => input.valueAsNumber === Number(input.max));
    for (const input of inputs) input.value = clear ? '0' : input.max;
    refreshUpgradePreview();
  };
  dialog.querySelector("[data-upgrade-submit]").onclick = confirmUpgradeFeed;
  if (typeof dialog.showModal === "function") dialog.showModal(); else dialog.setAttribute("open","");
  refreshUpgradePreview();
}

function selectedUpgradeMaterials(dialog) {
  const selected = {};
  for (const input of dialog.querySelectorAll("[data-feed-material]")) {
    const value = input.valueAsNumber, id = input.dataset.feedMaterial;
    if (!Number.isSafeInteger(value) || value < 0 || value > (SAVE.mats[id] || 0)) return null;
    if (value) selected[id] = value;
  }
  return selected;
}

function refreshUpgradePreview() {
  const dialog = activeUpgradePicker();
  if (!dialog) return;
  const own = SAVE.girls[dialog.dataset.upgradePicker], selected = selectedUpgradeMaterials(dialog), quote = selected && upgradeEstimate(own,selected);
  const inputs = [...dialog.querySelectorAll('[data-feed-material]')];
  for (const input of inputs) dialog.querySelector(`[data-feed-check="${input.dataset.feedMaterial}"]`).checked = input.valueAsNumber > 0;
  const toggle = dialog.querySelector('[data-feed-all]');
  toggle.disabled = inputs.every(input => !Number(input.max));
  toggle.textContent = inputs.every(input => input.valueAsNumber === Number(input.max)) ? '取消全选' : '选入全部';
  const summary = quote ? Object.entries(quote.spent).map(([id,count]) => `${atelierItem(id).name} ×${count}`).join("、") : "";
  dialog.querySelector("[data-feed-preview]").textContent = quote?.experience ? `预计 Lv.${own.level} → Lv.${quote.level} · 实得 ${quote.experience.toLocaleString()} 经验；将消耗 ${summary}` : selected ? "请选择至少一本经验材料。" : "用量须为库存范围内的整数。";
  dialog.querySelector("[data-upgrade-submit]").disabled = !quote?.experience || dialog.dataset.pending === "true";
}

async function confirmUpgradeFeed() {
  const dialog = activeUpgradePicker();
  if (!dialog || dialog.dataset.pending === "true") return;
  const id = dialog.dataset.upgradePicker, materials = selectedUpgradeMaterials(dialog), epoch = Store.sessionEpoch;
  if (!materials || !upgradeEstimate(SAVE.girls[id],materials).experience) return;
  dialog.dataset.pending = "true"; refreshUpgradePreview();
  try {
    const result = await Store.mutate("card_feed",{girl_id:id,materials});
    if (!result || epoch !== Store.sessionEpoch || !Store.ready()) return;
    if (dialog.isConnected) dialog.remove();
    refreshCultivationViews(id);
    if (currentWorkshop()?.id === id && currentWorkshop().tab === "upgrade") paintWorkshop("[data-upgrade-open]");
    toast((result.drops || []).join(" · ") || "经验已到账");
  } finally { delete dialog.dataset.pending; refreshUpgradePreview(); }
}
