const CIPHER_SYMBOLS = [["○","月"],["△","峰"],["□","门"],["◇","镜"],["×","岔"],["≋","潮"]];
let cipherPending = null;

function cipherDraft(view) {
  const board = SAVE?.cipher;
  if (!board || !Array.isArray(board.attempts)) return null;
  const tag = `${board.day}:${board.round}`, step = board.attempts.length+1;
  if (view.cipherDraft?.tag !== tag || view.cipherDraft.step !== step) view.cipherDraft = {tag,step,slot:0,code:["","","",""]};
  return view.cipherDraft;
}

function cipherView() {
  const view = currentWorkshop();
  return view?.tab === "dispatch" && view.dispatchRoom === "cipher" && view.epoch === Store.sessionEpoch ? view : null;
}

function cipherState(view) {
  const draft = cipherDraft(view), board = SAVE?.cipher;
  if (!draft) return null;
  const limit = board.round ? 4 : 5, over = board.solved || board.attempts.length >= limit;
  const code = draft.code.join(""), duplicate = board.attempts.some(entry => entry.code === code);
  return {draft,board,limit,over,code,duplicate,canGuess:!over && !duplicate && /^[0-5]{4}$/.test(code)};
}

function cipherPick(type, number) {
  const view = cipherView();
  if (!view || cipherPending?.epoch === view.epoch || !Number.isInteger(number)) return;
  const state = cipherState(view);
  if (!state || state.over) return;
  const {draft} = state;
  if (type === "slot" && number >= 0 && number < 4) draft.slot = number;
  else if (type === "symbol" && number >= 0 && number < CIPHER_SYMBOLS.length) {
    draft.code[draft.slot] = String(number);
    const blank = draft.code.indexOf("");
    draft.slot = blank < 0 ? (draft.slot+1)%4 : blank;
  } else return;
  paintWorkshop(`[data-cipher-${type}="${number}"]`);
}

async function cipherAction(action) {
  const view = cipherView();
  if (!view || cipherPending?.epoch === view.epoch) return;
  const state = cipherState(view);
  if (!state || !(action === "cipher_guess" && state.canGuess || action === "cipher_practice" && state.over)) return;
  const fields = {item_id:state.draft.tag};
  if (action === "cipher_guess") { fields.count = state.draft.step; fields.choice = state.code; }
  const token = {epoch:view.epoch};
  cipherPending = token;
  let answer;
  try { paintWorkshop(); answer = await Store.mutate(action,fields); }
  finally {
    if (cipherPending === token) cipherPending = null;
    if (Store.ready() && Store.sessionEpoch === token.epoch && cipherView() && document.querySelector('#modal-mask:not(.hidden) [data-cipher-panel]')) paintWorkshop();
  }
  if (answer && Store.ready() && Store.sessionEpoch === token.epoch) toast((answer.drops || []).join(" · ") || "译台已记录");
}

function cipherGlyph(index) {
  const sign = CIPHER_SYMBOLS[index];
  return sign ? `<i aria-hidden="true" class="cp-sign cp-${index}">${sign[0]}</i><span>${sign[1]}</span>` : `<i class="cp-sign" aria-hidden="true">·</i><span>空位</span>`;
}

function cipherWorkshopPanel(view) {
  const state = cipherState(view);
  if (!state) return `<section data-cipher-panel><h4>暗语译台</h4><p>暂无密本。</p></section>`;
  const {draft,board,limit,over} = state, locked = cipherPending?.epoch === Store.sessionEpoch;
  const ledger = board.attempts.map((entry,index) => `<li data-cipher-attempt="${index+1}"><small>${String(index+1).padStart(2,"0")}</small><div class="cp-code">${[...entry.code].map(key => `<b aria-label="${CIPHER_SYMBOLS[key][1]}">${CIPHER_SYMBOLS[key][0]}</b>`).join("")}</div><span>同位 <b>${entry.exact}</b> · 异位 <b>${entry.misplaced}</b></span></li>`).join("");
  return `<section class="cp-board" data-cipher-panel><header><div><small>${esc(board.day)} / ${board.round ? `练习 ${board.round}` : "日刊"}</small><h4>暗语译台</h4></div><strong>${board.solved ? "已译出" : over ? "已封卷" : `余 ${limit-board.attempts.length} 猜`}</strong></header><p>四符可重复。同位指符与位置都对；异位指符对、位置不对，每枚符只计一次。</p><ol class="cp-history" aria-label="译台反馈">${ledger || "<li class=\"cp-empty\">尚无猜测</li>"}</ol>${over ? `<button type="button" class="ws-command" data-cipher-action="cipher_practice" onclick="cipherAction('cipher_practice')" ${locked ? "disabled" : ""}>打开下一本练习 · 4 猜</button>` : `<div class="cp-slots" aria-label="本次猜测">${draft.code.map((value,index) => `<button type="button" data-cipher-slot="${index}" aria-pressed="${index === draft.slot}" aria-label="第 ${index+1} 位：${CIPHER_SYMBOLS[value]?.[1] || "空"}" onclick="cipherPick('slot',${index})" ${locked ? "disabled" : ""}>${cipherGlyph(value)}</button>`).join("")}</div><div class="cp-symbols" aria-label="可用符号">${CIPHER_SYMBOLS.map((sign,index) => `<button type="button" data-cipher-symbol="${index}" title="填入${sign[1]}" onclick="cipherPick('symbol',${index})" ${locked ? "disabled" : ""}>${cipherGlyph(index)}</button>`).join("")}</div><button type="button" class="ws-command" data-cipher-action="cipher_guess" onclick="cipherAction('cipher_guess')" ${!state.canGuess || locked ? "disabled" : ""}>${state.duplicate ? "本卷已猜过该序列" : `核对第 ${draft.step} 猜`}</button>`}<footer>${board.round ? "练习密本不发材料。" : "日刊首解：研训墨 ×2、星锻矿 ×1；北京时间零点换卷。"}</footer></section>`;
}
