const ATELIER_PLATES = {
  waymark: '<path class="leather" d="m15 6 17 5 16-5-6 29-10 14-11-14z"/><circle class="gold" cx="32" cy="36" r="19"/><circle class="shade" cx="32" cy="36" r="14"/><path class="shine" d="m32 23 4 9 9 4-9 4-4 9-4-9-9-4 9-4z"/>',
  lens: '<path class="leather" d="m10 45 19-18 8 8-18 19-9-1z"/><path class="gold" d="m16 39 7 9 4-4-7-9z"/><circle class="gold" cx="40" cy="23" r="17"/><circle class="shade" cx="40" cy="23" r="13"/><path class="shine" d="M30 24q2-13 16-11l-6 4q-6 1-10 7z"/>',
  reed: '<path class="gold" d="m6 44 40-32 10 2 2 7-43 33-8-2z"/><path class="leather" d="m12 47 39-31 3 4-40 30z"/><path class="tone" d="m30 35 6-5 4 17-4 6-3-5z"/><path class="line" d="m21 37 2 3m6-10 2 3m6-10 2 3"/>',
  banner: '<path class="gold" d="m15 6 5-2 4 51-6 3z"/><path class="tone" d="m20 10 16 2 19-4-5 13 5 9-17 5-16-4z"/><path class="shine" d="m35 15 3 6 7 2-6 3-2 5-3-5-7-2 6-3z"/>',
  stamp: '<path class="tone" d="m23 33 4-9q-9-7-3-14 8-8 16 1 4 8-5 14l5 9z"/><path class="gold" d="m10 39 24-10 19 6 2 12-25 10-20-8z"/><path class="leather" d="m13 42 19 6 20-9-1 6-20 9-18-7z"/>',
  dial: '<circle class="gold" cx="32" cy="32" r="26"/><circle class="shade" cx="32" cy="32" r="21"/><path class="paper" d="m48 16-10 23-22 10 10-23z"/><path class="tone" d="m48 16-10 23-6-7z"/><circle class="gold" cx="32" cy="32" r="4"/><path class="line" d="M32 10v5m0 34v5M10 32h5m34 0h5"/>',
  weapon: '<path class="ink" d="m50 5 7-1-2 11-23 25-6-6z"/><path class="steel" d="m51 7 4-1-4 9-20 22-3-3z"/><path class="shine" d="m51 9-21 25 3-1z"/><path class="gold" d="m20 30 3-3 16 15-4 5-4-9-5-3z"/><path class="leather" d="m23 36 5 4-13 15-5-5z"/><path class="line" d="m14 47 6 2m-3-6 6 2"/><path class="gold" d="m8 50 8 6-4 4-7-6z"/>',
  head: '<path class="leather" d="M20 12Q32-3 49 8l-8 1 8 5-12 3-13 9z"/><path class="ink" d="M10 37q0-22 22-24 20 0 23 20l-4 18-10 6-8-13-5 15-14-10z"/><path class="steel" d="M13 33q2-18 19-18 18 1 20 16l-14 6-5 17-4-16z"/><path class="shade" d="m15 35 13 6-3 13-9-7zm25 5 11-5-3 13-7 5z"/><path class="gold" d="m13 31 16 3 4-17 4 17 15-6 1 4-17 7-3 12-3-12-17-4z"/><path class="shine" d="m18 27 9-9 3 1-4 8z"/>',
  hands: '<g transform="rotate(-21 32 32)"><path class="ink" d="m14 42 4-14 2-14 5-3 4 3 1-7 5-2 4 4 2-3 5 3v7l4-2 4 5-4 17-7 7-2 13-25-1z"/><path class="steel" d="m19 38 3-15 8-6 16 3 2 13-10 12-14-1z"/><path class="shade" d="M21 21v-7l4-2 3 3-1 7zm10-5 1-7 4-2 3 3-1 8zm10 2 1-9 4 2-1 8zm7 5 2-7 3 3-2 6z"/><path class="gold" d="m19 42 22 3-1 9-23-1z"/><path class="leather" d="m19 48 20 2-1 6-21-2z"/><path class="shine" d="m25 28 10-7 6 4-8 4-5 9z"/></g>',
  body: '<path class="leather" d="M13 19 9 52l13 3 9-7 10 10 13-6-7-37z"/><path class="ink" d="m13 13 9-4 8 6 6-1 7-6 10 7 6 12-10 6-6 22-12 5-14-8-3-19-11-6z"/><path class="steel" d="m20 14 11 5 12-6 4 12-4 17-13 6-12-8-3-16z"/><path class="shade" d="m13 15 6-2-5 15-9-3zm32-2 6 4 6 9-10 3zM19 42l12 7 12-6-2 8-11 6-10-7z"/><path class="gold" d="m20 14 11 5 13-6-2 5-11 6-12-6zm-2 20 12 5 14-7-3 6-11 6-11-6z"/><path class="shine" d="m20 23 8 5-2 6-6-3zm16 3 6-5-1 8-5 3z"/>',
  feet: '<path class="ink" d="m25 7 24 4-6 25 1 6-6 8-22 7-11-3v-7l16-8 2-12z"/><path class="leather" d="m27 10 19 3-6 23-12 5-5-4 4-12zM9 47l16-7 5 3 11-6-1 8-23 9-9-2z"/><path class="steel" d="m24 15 23 3-2 6-21-4zm-1 9 20 3-2 6-19-5zm-9 21 8-4 6 5-11 6-9-2z"/><path class="gold" d="m22 9 26 4 1-5-24-4zm0 24 6 3 12-3-1 5-11 4-7-5z"/><path class="shade" d="m7 51 11 2 24-9-1 5-24 9-11-3z"/>',
  crystal: '<path class="ink" d="m8 36 7-14 10 6L33 5l11 13v13l9-5 4 16-14 12-22 3z"/><path class="tone" d="m25 34 9-24 7 10-4 28-8 6zM11 36l5-9 10 5 1 20-9-5zm30 5 11-10 2 11-14 10z"/><path class="shine" d="m33 13-4 19 7-7zm-18 18-1 8 7 4-2-9zm32 6-5 9 8-5z"/>',
  ore: '<path class="ink" d="m4 43 6-18 13-3L33 9l17 6 10 22-4 10-25 8-18-3z"/><path class="shade" d="m7 42 10-10 8 1 6 17-16-2zm21-10 17-9 10 12-4 10-17 5z"/><path class="steel" d="m12 27 12-2 6 6-9 6-10-1zm14-5 8-10 13 5-6 6-11 6z"/><path class="gold" d="m37 16 4 2-8 5 2 4-10 5 7-9zM43 30l5-4 2 3-5 4 2 4-10 5 7-8z"/>',
  core: '<path class="gold" d="m31 3 22 14 5 23-23 20-26-12-4-22z"/><path class="ink" d="m31 8 17 12 5 18-19 17-20-9-4-18z"/><path class="tone" d="m31 13 12 12-4 20-9 5-11-13z"/><path class="shine" d="m31 18 5 9-5 11-3-7z"/><path class="gold" d="m8 29 9-5 5 3-8 4zm36 12 8-8 2 4-8 8z"/>',
  manual: '<path class="ink" d="m7 17 33-9 17 8 1 32-36 10-15-9z"/><path class="paper" d="m11 44 28-9 15 7-31 11-12-5z"/><path class="tone" d="m9 17 30-7 16 7-32 10zM10 20l11 7v22l-11-6z"/><path class="shade" d="m24 29 30-10v21l-30 11z"/><path class="gold" d="m35 23 6-2v15l-3-1-3 4zM25 13l3-1 12 6-4 1z"/><path class="line" d="m25 51 27-9"/>',
  ticket: '<g transform="rotate(-15 32 32)"><path class="ink" d="M5 15h54v11a6 6 0 0 0 0 12v11H5V38a6 6 0 0 0 0-12z"/><path class="paper" d="M9 19h46v5q-10 8 0 16v5H9v-5q10-8 0-16z"/><path class="tone" d="M15 22h23v20H15q7-9 0-20z"/><path class="gold" d="m26 24 2 5 5 2-5 3-2 5-2-5-5-3 5-2z"/><path class="line" d="M42 20v4m0 3v4m0 3v4m0 3v3m4-18h5m-5 6h4m-4 5h5"/></g>',
  dust: '<path class="leather" d="m22 11 9 4 11-5-4 12 10 12q15 19-13 23-29 1-24-16l13-19z"/><path class="tone" d="m25 29 12-3 7 12q8 15-12 16-20 0-16-10z"/><path class="gold" d="m20 20 19-1 2 5-19 2z"/><path class="shine" d="m27 36 2-4 2 4 4 2-4 2-2 4-2-4-4-2zm11 9 1-3 2 3-2 3z"/>',
  essence: '<path class="ink" d="M23 7h18v6l-3 2v12l12 14q9 16-17 17-27-2-20-16l13-16V15l-3-2z"/><path class="steel" d="M28 14h7v13L20 43q-7 10 13 12 20-1 13-12L33 28V16h-5z"/><path class="tone" d="m19 41 11 2 12-5 5 9q-2 8-15 7-17-1-14-9z"/><path class="gold" d="M24 7h15v6H24z"/><path class="shine" d="m25 33-6 9 3 1 7-10z"/>',
  ingot: '<path class="ink" d="m5 40 8-15 31-9 13 7 3 17-38 13z"/><path class="steel" d="m14 26 30-8 10 6-32 10z"/><path class="shade" d="m8 40 6-11 7 8-1 11zm15-3 31-10 3 11-34 11z"/><path class="gold" d="m23 27 12-4 8 2-12 5z"/>',
  thread: '<path class="ink" d="m13 14 23-6 12 7-5 6 4 19 5 2 1 7-27 7-14-7 3-7 4-2-4-18-4-1z"/><path class="leather" d="m15 15 22-5 8 5-23 6zm1 29 23-6 10 6-23 6z"/><path class="tone" d="m19 23 21-5 5 21-22 7z"/><path class="line" d="m21 28 19-6m-18 11 20-6m-19 11 20-6m-17 9 18-5"/><path class="gold" d="m40 46 5 5 10-4-1 3-10 5-7-7z"/>',
  rivet: '<path class="ink" d="m9 16 12-7 11 5 1 9-7 5 16 18-3 7-6-2-16-19-8-3zm26-1 10-4 9 7-1 7-5 2 6 15-2 8-6-5-6-17-6-5z"/><path class="steel" d="m11 17 11-5 7 4-9 7zm9 12 5-3 15 19-4 3zm16-12 8-4 7 5-8 4zm8 11 4-1 6 17-4-2z"/><path class="gold" d="m14 17 8-3 3 2-7 3zm24 1 6-3 4 2-6 2z"/>',
  seal: '<path class="tone" d="m15 7 16 8L47 6l2 16-10 8 7 17-12-3-9 12-4-19-14 2 10-14z"/><path class="gold" d="m32 14 11 10-4 13-14 1-8-12z"/><path class="ink" d="m32 18 6 8-4 7-10-4z"/><path class="shine" d="m31 21 2 4 3 2-5 1-2 2 1-5-2-2z"/>',
  scroll: '<path class="ink" d="m15 6 36 3 3 9-5 4-3 28-10 8-29-5-3-11 6-4z"/><path class="paper" d="m17 12 27 3-4 28-5 8-22-3 5-7z"/><path class="gold" d="m13 8 35 2 4 5-4 3-33-4zm-5 34 27 5 4 6-4 2-27-6z"/><path class="line" d="m23 22 14 2m-15 4 11 2m-12 4 13 2"/>',
  gem: '<path class="ink" d="M4 24 17 9l29 1 14 15-28 33z"/><path class="tone" d="m8 25 11-12 24 1 13 12-24 27z"/><path class="shine" d="m12 23 9-8h8l-8 9zm13 4 10-1-3 21z"/><path class="line" d="m8 25 48 1m-34-2 9-11 11 12-10 27z"/>',
  chip: '<path class="gold" d="m12 19 8-12 6 3 3-7 6 4 7-1 3 7 9 2-1 7 7 5-4 7 2 7-8 4-3 9-8-1-6 6-6-4-9 1-3-8-8-2 1-8-6-5 4-7-1-7z"/><path class="ink" d="m19 15 23-1 12 18-12 18-22 1L9 33z"/><path class="tone" d="m22 21 16-2 9 13-9 13-16-1-7-11z"/><path class="trace" d="m16 26 7 3m22 8-7-2m-12 10 1-7m7-18-1 7"/>',
  shard: '<path class="ink" d="m12 19 24-10 16 8-6 11 5 10-20 19-18-6 4-14-7-8z"/><path class="tone" d="m17 20 18-8 12 6-7 12 5 7-16 16-12-4 5-15-7-8z"/><path class="shine" d="m22 22 14-6-9 12 3 6-9 12 4-14-5-6z"/>'
};

function atelierItem(id) { return Store.profile?.items?.items?.[id]; }

const ATELIER_GEM_CUTS = {
  ruby: '<path class="ink" d="M6 25 19 10h25l14 15-25 33z"/><path class="tone" d="m10 25 11-12h21l12 12-21 28z"/><path class="shade" d="m17 26 10-12-4 11 10 25-19-25zm17 24 7-24h12z"/><path class="shine" d="m13 24 10-9h7l-7 9zm15 3h9l-4 16z"/><path class="line" d="m10 25 44 1m-31-1 9-11 9 12-8 27z"/>',
  jade: '<path class="ink" d="M13 47C1 32 18 9 34 6c14-3 26 7 25 19-2 17-16 30-31 32-7 1-12-4-15-10z"/><path class="tone" d="M16 46C6 33 20 13 34 10c13-3 22 6 21 16-1 14-15 26-27 27-5 1-9-2-12-7z"/><path class="shade" d="M52 23c-5 15-18 26-32 25 4 6 12 7 21 1 10-7 15-18 11-26z"/><path class="shine" d="M16 33q1-14 19-19l-6 6q-7 3-13 13z"/><path class="line" d="M17 40q4-8 13-13t12-11m-13 13 1 8-4 6"/>',
  sapphire: '<path class="ink" d="m9 19 12-11h25l11 11v26L46 56H18L7 46z"/><path class="tone" d="m13 20 10-9h21l10 10v22l-10 10H20L11 43z"/><path class="shade" d="m12 43 8 10h24l9-10-7 1-6 4H24l-5-4z"/><path class="line" d="m19 23 7-7h15l7 8-1 16-7 7H25l-7-8zm5 3h17v12H24z"/><path class="shine" d="m14 21 10-8h11l-4 4h-5l-8 8z"/>',
  wind: '<path class="ink" d="M56 5c-2 12-5 29-13 40-6 9-15 15-26 11C2 50 8 35 19 27c14-9 25-10 37-22z"/><path class="tone" d="M50 15c-3 13-5 22-12 30-5 6-13 11-20 8-9-4-8-14 3-22 9-7 19-10 29-16z"/><path class="shade" d="m43 28-8 17-11 6 7-9-13 5z"/><path class="shine" d="M14 42q2-11 21-17l-7 7-5 2z"/><path class="line" d="M18 50q12-7 28-27m-17 13 1-5m4 7 8-5"/>',
  amethyst: '<path class="ink" d="m16 48-6-14 7-15 6 6 8-22 14 10-1 17 7-9 7 7-8 22-17 10z"/><path class="tone" d="m27 27 5-19 10 8-4 26-7 13-8-9zM14 35l4-10 8 8-4 14zm29 2 7-11 4 4-7 17-10 7z"/><path class="shade" d="m31 20 9-6-4 28-5 10-3-10zm16 18 6-7-7 15-6 5z"/><path class="shine" d="m32 10-5 15 6-6 4-5zM16 34l3-5 3 5-4 5z"/><path class="line" d="m33 20-5 23m21-9-5 10"/>'
};

const ATELIER_FAMILY_INKS = {
  edge: 'M36 19 24 33h7l-4 11 15-20-9 5z', aegis: 'm32 21 9 4-1 9-8 8-8-8-1-9zm0 6v10', tempo: 'm20 33 6-1 4-8 4 14 4-8h6',
  ember: 'M31 36q-1 4-4 7-3 5 4 7 9-1 6-8l-3 2z', tide: 'M24 43q4-4 8 0t8 0m-16 5q4-4 8 0t8 0', grove: 'M25 49q-2-11 13-12-1 13-13 12zm0 0 9-8',
  astral: 'm31 36 2 6 6 2-6 2-2 6-2-6-6-2 6-2z', umbra: 'M33 37a7 7 0 1 0 6 11q-12 1-6-11z', chrono: 'M25 37h12m-10 1q0 4 4 6-4 2-4 6m9-12q0 4-5 6 5 2 5 6m-11 1h12'
};

const ATELIER_GRADE_MARKS = [
  'M12 3a9 9 0 1 0 .1 0M12 10v3',
  'M12 16C5 16 3 11 3 7q9 1 9 9zm0 0Q12 7 21 6q-1 10-9 10zm0 0v5m0-19 1 2 2 1-2 1-1 2-1-2-2-1 2-1z',
  'M6 6C13 0 24 8 20 15m-2 3C11 24 0 16 4 9M5 13q7 5 15-6m-15 4-2-3m9 3v2',
  'm12 2 8 10-8 10-8-10zm0 5 2 4 4 1-4 2-2 4-2-4-4-2 4-1zM2 7l3 2m17-2-3 2M2 17l3-2m17 2-3-2',
  'M2 7l6 4 4-8 4 8 6-4-3 12H5zm3 14h14m-7-9 2 2-2 2-2-2z',
  'M12 6a6 6 0 1 1-5.2 3M12 2v2m8 1-2 3m4 5-3-1m-1 7-2-2m-5 6 1-3m-8-2 3-2m-5-5 3 1',
  'M1 6l7 3-2 3 3 4-6-3zm22 0-7 3 2 3-3 4 6-3zM12 2l1.5 4 3.5-1-2 3.5 3 2.5-4 .5-2 4.5-1.5-4.5L7 11l2.5-2.5L8 5l3 1zm0 15 2 2-2 3-2-3z'
];

function atelierIcon(id, quality, shape) {
  const item = atelierItem(id), kind = shape || item?.icon || id;
  const icon = ({ helm:"head", gauntlet:"hands", armor:"body", boots:"feet" })[kind] || kind;
  const tier = quality || item?.quality || "white", grades = Store.profile?.items?.qualities;
  const grade = grades?.[tier], paint = /^#[0-9a-f]{6}$/i.test(grade?.color || "") ? grade.color : "#657886";
  const [group, position] = id.split("_"), family = icon === "chip" ? group : icon === "essence" ? position : "";
  const accent = GEMS[id]?.color || ({edge:"#e58d6f", aegis:"#87bfa0", tempo:"#87b2dc"})[family] || MATS[id]?.color || paint;
  const rank = Math.min(6, Math.max(0, Number(grade?.rank) || 0));
  const plate = (icon === "gem" && ATELIER_GEM_CUTS[id]) || ATELIER_PLATES[icon] || ATELIER_PLATES.seal;
  const ink = ATELIER_FAMILY_INKS[family], label = [item?.name, grade?.name].filter(Boolean).join(" · ");
  const chipNumber = icon === "chip" && /^[1-6]$/.test(position) ? `<i class="at-position" aria-hidden="true">${position}</i>` : "";
  const markings = grade ? `<svg class="at-emblem" viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="${ATELIER_GRADE_MARKS[rank]}"/></svg><small class="at-ribbon" aria-hidden="true">${esc(grade.name.split(" · ").at(-1))}</small><i class="at-pips" aria-hidden="true" style="--marks:${rank + 1}"></i>` : "";
  return `<span class="at-icon" data-rarity-quality="${esc(tier)}" data-grade-rank="${rank}" role="img" aria-label="${esc(label || id)}" style="--quality:${paint};--item-color:${accent}" title="${esc([label, grade?.emblem].filter(Boolean).join(' · '))}"><svg class="at-plate" viewBox="0 0 64 64" aria-hidden="true" focusable="false"><g stroke-linecap="round" stroke-linejoin="round">${plate}${ink ? `<path class="at-family-ink" d="${ink}"/>` : ""}</g></svg>${chipNumber}${markings}</span>`;
}

function atelierStarName(star) { return star >= 12 ? "冠冕" : star >= 9 ? "绯红" : star >= 6 ? "黄金" : ""; }

function atelierCrest(star) {
  const name = atelierStarName(star);
  return name ? `<span class="at-crest" data-star-prestige="${star >= 12 ? "crown" : star >= 9 ? "crimson" : "gold"}"><svg viewBox="0 0 24 18" aria-hidden="true"><path d="m2 4 6 4 4-7 4 7 6-4-4 10H6zM6 15h12v2H6z"/></svg>${name} · ${star}星</span>` : "";
}

function atelierMilestones(id) {
  const milestones = typeof starMilestones === "function" ? starMilestones(id) : [];
  if (!milestones.length) return "";
  const first = milestones.find(entry => !entry.active);
  return `<section class="at-route"><h4>星级效果路线 <small>${first ? `下一档 · ${first.star} 星` : "全档生效"}</small></h4><ol>${milestones.map(entry => `<li data-star-milestone="${entry.star}" class="${entry.active ? "is-active" : entry === first ? "is-next" : ""}"><b>${entry.star} 星 · ${esc(entry.name)}</b><small>${esc(entry.description)}</small><em>${entry.active ? "已生效" : entry === first ? "下一档" : "未解锁"}</em></li>`).join("")}</ol></section>`;
}

let atelierGearPending = null;

function atelierGearGrade(level, rules = Store.profile?.equipment) {
  const steps = rules?.quality_steps || [{level:0,quality:"white"}];
  const step = steps.findLast(entry => level >= entry.level) || steps[0], next = steps.find(entry => entry.level > level);
  return { quality:step.quality, glow:next ? Math.min(2,Math.max(0,level - step.level)) : 2, next };
}

function atelierGearArt(quote) {
  const { quality, glow } = quote.grade;
  const prestige = quote.star >= 15 ? "zenith" : quote.star > 12 ? "beyond" : quote.star === 12 ? "crown" : "";
  return `<span class="at-gear-aura" data-gear-quality="${esc(quality)}" data-gear-glow="${glow}" data-gear-prestige="${prestige}" style="--gear-light:var(--grade-${esc(quality)})">${atelierIcon(`gear_${quote.slot.id}`,quality,quote.slot.id)}${prestige ? `<i class="at-gear-crown" aria-hidden="true">${prestige === "zenith" ? "✷" : "♛"}</i>` : ""}</span>`;
}

function atelierGearStat(quote, level, star) {
  const levels = level * quote.perLevel, stars = star * quote.slot.per_star;
  return quote.slot.stat === "signature" && quote.stat !== "spd" ? Math.round(((1 + levels / 100) * (1 + stars / 100) - 1) * 10000) / 100 : levels + stars;
}

function atelierGearQuote(girlId, slotId) {
  const rules = Store.profile?.equipment, own = SAVE?.girls[girlId], slot = rules?.slots.find(entry => entry.id === slotId);
  if (!own?.star || !slot) return null;
  const piece = own.equipment?.[slot.id], signature = slot.stat === "signature" ? cultivationRules()?.weapons[girlId] : null;
  const level = (slot.id === "weapon" ? own.weapon_level : piece?.level) || 0, star = piece?.star || 0;
  const levelCost = rules.level_costs[level] || 0, starCost = rules.star_costs[star] || 0;
  const requiredLevel = rules.star_levels[star] || 0, requiredStar = rules.star_card_stars[star] || 0;
  const levelStock = SAVE.mats[slot.level_mat] || 0, starStock = SAVE.mats[rules.star_mat] || 0;
  const markCost = rules.star_mark_costs?.[star] || 0, markStock = SAVE.mats[rules.star_mark_mat] || 0;
  let autoCost = 0, autoLevels = 0;
  for (const cost of rules.level_costs.slice(level)) {
    if (cost <= 0 || autoCost + cost > levelStock) break;
    autoCost += cost; autoLevels++;
  }
  return { slot, name:signature?.name || slot.name, stat:signature?.stat || slot.stat, perLevel:signature ? cultivationRules().weapon_per_level : slot.per_level,
    level, star, grade:atelierGearGrade(level,rules), levelCost, autoLevels, autoCost, starCost, requiredLevel, requiredStar, levelStock, starStock, markCost, markStock,
    canUpgrade:levelCost > 0 && levelStock >= levelCost, canAuto:autoLevels > 0, canStar:starCost > 0 && starStock >= starCost && markStock >= markCost && level >= requiredLevel && own.star >= requiredStar };
}

function selectAtelierGear(slot) {
  const view = currentWorkshop();
  if (!view || view.tab !== "weapon" || !atelierGearQuote(view.id, slot)) return;
  view.gear = slot;
  paintWorkshop(`[data-atelier-gear="${slot}"]`);
}

function openGrowthGear(id, slot) {
  if (!requireGame() || !workshopRules() || !atelierGearQuote(id,slot)) return;
  openWorkshop(id,"weapon");
  selectAtelierGear(slot);
}

async function actAtelierGear(action) {
  const view = currentWorkshop(), flags = { gear_upgrade:"canUpgrade", gear_auto:"canAuto", gear_star:"canStar" };
  if (!view || view.tab !== "weapon" || !Object.hasOwn(flags,action) || atelierGearPending?.epoch === Store.sessionEpoch) return;
  const slot = view.gear || "weapon", quote = atelierGearQuote(view.id,slot);
  if (!quote?.[flags[action]]) return;
  const token = { epoch:view.epoch };
  atelierGearPending = token;
  let result;
  try {
    paintWorkshop();
    result = await Store.mutate(action, { girl_id:view.id, item_id:slot });
  } finally {
    if (atelierGearPending === token) atelierGearPending = null;
    if (token.epoch === Store.sessionEpoch && Store.ready()) {
      if (result) refreshCultivationViews(view.id);
      const open = document.querySelector("#modal-mask:not(.hidden) .hero-workshop");
      if (open?.dataset.tab === "weapon" && (!result || open.dataset.cultivationCard !== view.id)) paintWorkshop();
    }
  }
  if (result && token.epoch === Store.sessionEpoch && Store.ready()) toast((result.drops || []).join(" · ") || "装备培养已确认");
}

function atelierGearSets(own) {
  const rules = Store.profile.equipment;
  const tiers = rules.set_tiers.map(tier => ({ ...tier, count:rules.slots.filter(slot => (own.equipment?.[slot.id]?.star || 0) >= tier.star).length }));
  const next = tiers.find(tier => tier.count < tier.pieces);
  return `<section class="at-set-ledger"><h4>星纹套装 <small>${tiers.filter(tier => tier.count >= tier.pieces).length}/${tiers.length} 档已生效</small></h4><ol>${tiers.map(tier => `<li data-gear-set="${tier.pieces}" class="${tier.count >= tier.pieces ? "is-active" : tier === next ? "is-next" : ""}"><b>${esc(tier.name)}</b><small>${tier.star} 星装备 · ${tier.count}/${tier.pieces} 件</small><em>${tier.count >= tier.pieces ? "已生效" : tier === next ? "下一档" : "未解锁"}</em><span>${Object.entries(tier.stats).map(([stat,value]) => workshopEffect({stat,value})).join(" · ")}</span></li>`).join("")}</ol></section>`;
}

function atelierRack(girlId, selected) {
  const rules = Store.profile.equipment;
  return `<div class="at-rack"><svg class="at-stand" viewBox="0 0 240 320" aria-hidden="true" focusable="false"><g fill="none" stroke="currentColor"><path d="M120 12v292M19 128h202M35 255h170" stroke-dasharray="2 5"/><path d="M87 48q32-36 65 0l-5 25-18 14-22-8-16-13zm11 35-26 10-20 30 14 36 15-7 2 42 24 19-13 77 15 6 16-57 8 56 17-4-4-77 21-19 1-41 15 7 10-36-22-32-32-11-14 10z"/><path d="m75 99 20 24 23 8 24-6 23-25m-82 58 34 13 44-12m-77 28 34 15 41-13m-52-94 10 103m14-107-11 108m-12 12 12 18 15-17M79 297h76l10 7H71z"/></g></svg>${rules.slots.map(slot => {
    const quote = atelierGearQuote(girlId,slot.id);
    return `<button type="button" class="at-rack-piece" data-atelier-gear="${slot.id}" aria-label="${esc(quote.name)}，强化 ${quote.level}，装备 ${quote.star} 星" aria-pressed="${selected === slot.id}" onclick="selectAtelierGear('${slot.id}')">${atelierGearArt(quote)}<span>${esc(slot.name)}</span><small>+${quote.level} · ${quote.star}星</small></button>`;
  }).join("")}</div>`;
}

function atelierBoundWeapon(girlId, own) {
  const rules = cultivationRules(), rune = rules.engravings[own.weapon_rune], options = cultivationOptions(girlId);
  return `<section class="at-signature"><h4>专武觉醒 <small>${own.weapon_awake || 0}/${rules.weapon_cores.length} 阶</small></h4><p>每阶生命 +${rules.weapon_vitality_per_awake}%、初始怒气 +${rules.weapon_rage_per_awake}</p><div class="cultivation-buttons">${cultivationButton(girlId,options.find(option => option.action === "weapon_awaken"))}</div><h4>铭刻 <small>${rune ? esc(rune.name) : "未铭刻"}</small></h4><p>专武 +${rules.weapon_rune_level} 开启 · 每次 ${rules.weapon_rune_ore} 星锻矿（持有 ${SAVE.mats.forge_ore || 0}） · 新铭纹替换旧铭纹</p><div class="cultivation-buttons">${options.filter(option => option.action === "weapon_engrave").map(option => cultivationButton(girlId,option)).join("")}</div></section>`;
}

function atelierGearPanel(view, girl, own) {
  const rules = Store.profile.equipment, quote = atelierGearQuote(girl.id,view.gear || "weapon");
  const { slot, level, star, levelStock, levelCost, starStock, starCost, perLevel } = quote;
  const item = atelierItem(`gear_${slot.id}`), title = esc(quote.name), stat = WORKSHOP_STATS[quote.stat], unit = quote.stat === "spd" ? "" : "%";
  const grades = Store.profile.items?.qualities || {}, grade = grades[quote.grade.quality], nextGrade = quote.grade.next;
  const levelMat = atelierItem(slot.level_mat)?.name || rules.materials[slot.level_mat]?.name || CULTIVATION_ITEMS[slot.level_mat]?.name || slot.level_mat;
  const starMat = atelierItem(rules.star_mat)?.name || rules.materials[rules.star_mat]?.name || rules.star_mat;
  const current = atelierGearStat(quote,level,star);
  const controls = [
    ["gear_upgrade", levelCost ? `强化至 +${level + 1}` : "已满级", levelCost ? `${levelMat} ${levelStock}/${levelCost}` : `强化 +${level}`,quote.canUpgrade],
    ["gear_auto", "一键强化", quote.canAuto ? `至 +${level + quote.autoLevels} · 共 ${quote.autoCost} ${levelMat}` : levelCost ? "材料不足一档" : "已达到强化上限",quote.canAuto],
    ["gear_star", starCost ? `${star >= 12 ? "超限至" : "突破至"} ${star + 1} 星` : "终阶已满星", starCost ? `${starMat} ${starStock}/${starCost}${quote.markCost ? ` · ${atelierItem(rules.star_mark_mat)?.name || "行旅徽记"} ${quote.markStock}/${quote.markCost}` : ""}` : `装备 ${star} 星`,quote.canStar]
  ];
  return `<div class="at-equip-layout"><aside class="at-owner"><small class="at-label">${esc(girl.name)} · 绑定装备</small>${atelierGearSets(own)}</aside><section class="at-blueprint"><div class="at-blueprint-title"><small>星霜工坊</small><h4>五部位装配图</h4></div>${atelierRack(girl.id,slot.id)}<div class="at-plate-caption"><span>装配图 ${girl.id.slice(1)}</span><b>固定装备 · 伍件</b></div></section><section class="at-gear-info"><small class="at-label">${esc(slot.name)} · 装备档案</small>
    <div class="at-object-title">${atelierGearArt(quote)}<div><h4>${title}</h4><p>${grade ? `${esc(grade.name)} · ` : ""}环光 ${quote.grade.glow} · ${star === 15 ? "十五星终阶" : star > 12 ? `超限 ${star} 星` : star === 12 ? "十二星冠" : "角色绑定"}</p></div></div>
    <dl class="at-readout"><div><dt>强化等级</dt><dd>+${level} / ${rules.level_costs.length}</dd></div><div><dt>装备星级</dt><dd>${star} / ${rules.star_costs.length} 星</dd></div><div><dt>当前${stat}</dt><dd>+${current}${unit}</dd></div></dl>
    <div class="at-forecast">${nextGrade ? `<span>下一品质</span><b>+${nextGrade.level} · ${esc(grades[nextGrade.quality]?.name || nextGrade.quality)}</b>` : "<span>终阶品质</span>"}${levelCost ? `<span>强化下一档</span><b>${stat} +${current}${unit} → +${atelierGearStat(quote,level+1,star)}${unit}</b>` : "<span>强化已满</span>"}${starCost ? `<span>突破下一星</span><b>${stat} +${current}${unit} → +${atelierGearStat(quote,level,star+1)}${unit}</b>` : "<span>突破已满</span>"}</div>
    ${starCost ? `<p class="at-requirements">突破条件：装备强化 +${level}/${quote.requiredLevel} · 角色 ${own.star}/${quote.requiredStar} 星</p>` : ""}<div class="at-gear-actions">${controls.map(([action,label,cost,ready]) => `<button type="button" class="ws-command" data-gear-action="${action}" ${ready && atelierGearPending?.epoch !== view.epoch ? "" : "disabled"} onclick="actAtelierGear('${action}')"><b>${esc(label)}</b><small>${esc(cost)}</small></button>`).join("")}</div>${item ? `<p class="at-source">来源：${esc(item.source)}<br>用途：${esc(item.use)}</p>` : ""}</section>${slot.id === "weapon" ? atelierBoundWeapon(girl.id,own) : ""}</div>`;
}

function atelierItemDetails(id) {
  const identity = artifactIdentity(id), item = atelierItem(identity?.base || id);
  if (!item) return "";
  const refine = item.category === "gem" || item.category === "chip" ? refinementItem(item.category === "gem" ? "gems" : "chips",id) : null;
  if (refine?.quality) {
    const grade = Store.profile.artifact_grades.qualities[refine.grade];
    const baseEffect = identity.base === "wind" ? item.effect.split("。").slice(1).join("。") : item.effect;
    return `<dl class="at-item-notes"><dt>品质</dt><dd data-artifact-quality="${grade.id}">${esc(grade.name)} · Lv.${refine.level} / ${refine.rank}阶</dd><dt>用途</dt><dd>${esc(item.use)}</dd><dt>效果</dt><dd>${esc(baseEffect)} 固有${WORKSHOP_STATS[refine.spec.stat]} +${refine.baseValue}，当前 +${refine.value}${refine.extraHp ? `，另加生命 ${refine.extraHp}` : ""}。</dd><dt>词条</dt><dd>${workshopGradeEffects(refine)}</dd><dt>来源</dt><dd>${esc(item.source)} 召唤品质概率 ${grade.weight/100}%。</dd></dl>`;
  }
  const upgrade = refine ? `<dt>精炼</dt><dd>${WORKSHOP_STATS[refine.spec.stat]}基础 +${refine.spec.base}，每级 +${refine.spec.step}；${refine.spec.effects.map(effect => `+${effect.level} ${esc(effect.name)}`).join("、")}。</dd>` : "";
  return `<dl class="at-item-notes"><dt>用途</dt><dd>${esc(item.use)}</dd><dt>效果</dt><dd>${esc(item.effect)}</dd>${upgrade}<dt>来源</dt><dd>${esc(item.source)}</dd></dl>`;
}

function atelierMaterialShelf(group) {
  const entries = Object.entries(Store.profile.items.items).filter(([id,item]) => {
    if (group === "workshop") return item.category === "forge" || item.category === "experience" || id === "training_manual";
    if (group === "stars") return item.category === "legacy" || item.category === "growth" && id !== "training_manual";
    return item.category === ({tickets:"ticket",sigils:"sigil"})[group];
  });
  const cost = cultivationRules()?.shards?.legacy_dust_per_ore;
  return `<div class="at-shelf">${entries.map(([id,item]) => `<article data-bag-material="${id}"><div class="at-object-title">${atelierIcon(id)}<div><small class="at-label">${esc(Store.profile.items.categories[item.category].name)}</small><h4>${esc(item.name)}</h4><strong>持有 ${SAVE.mats[id] || 0}</strong></div></div>${atelierItemDetails(id)}${item.category === "legacy" && cost ? `<button type="button" class="ws-command" data-dust-exchange="${id}" onclick="exchangeLegacyDust('${id}')" ${(SAVE.mats[id] || 0) >= cost ? "" : "disabled"}>${cost} 旧尘换 1 星锻矿</button>` : ""}</article>`).join("")}</div>`;
}

function atelierEmptySocket(tab, state) {
  const closed = state === "is-closed";
  const path = closed ? '<path d="M8 12V9a4 4 0 1 1 8 0v3M6 12h12v8H6zM12 15v2"/>' : tab === "chips" ? '<path d="m8 4 8 0 5 8-5 8H8l-5-8zM8 12h8m-4-4v8"/>' : tab === "gems" ? '<path d="m3 9 5-5h8l5 5-9 11zM4 9h16M8 5l4 15 4-15"/>' : '<path d="m12 2 3 7 7 3-7 3-3 7-3-7-7-3 7-3zM12 7v10M7 12h10"/>';
  return `<svg class="at-socket-mark" viewBox="0 0 24 24" aria-hidden="true">${path}</svg>`;
}

function atelierGrowthDossier(id, goal) {
  const girl = GIRLS.find(entry => entry.id === id), own = SAVE.girls[id], rules = Store.profile.equipment;
  const stats = finalStats(id), milestones = typeof starMilestones === "function" ? starMilestones(id) : [];
  const next = milestones.find(tier => !tier.active), bonds = typeof collectionBonds === "function" ? collectionBonds(id) : [];
  const options = workshopTabKeys().map(key => `<button type="button" data-open-workshop="${key}" onclick="openWorkshop('${id}','${key}')">${esc(workshopTabName(key))}<span aria-hidden="true">›</span></button>`).join("");
  return `<article class="at-dossier" data-growth-dossier="${id}" id="growth-cultivation"><header class="at-dossier-head"><h3>星霜培养手册</h3><label class="growth-select">切换角色<select id="growth-hero" onchange="selectGrowthGirl(this.value)">${growthHeroOptions(id,true)}</select></label></header><div class="at-dossier-body"><section class="at-profile"><div class="at-profile-head" ${typeof trainingFrame === "function" ? trainingFrame(own) : ""}><img src="${girl.img}" alt="${esc(girl.name)}"><div><small>${girl.rarity} · ${ELEM[girl.elem].name}系</small><h4>${esc(girl.name)}</h4><span>Lv.${own.level} · ${own.star} 星 · 觉醒 ${own.awakening || 0} 阶</span>${typeof trainingGradeTag === "function" ? trainingGradeTag(own) : ""}${atelierCrest(own.star)}</div></div><dl class="at-profile-stats">${Object.entries({hp:"生命",atk:"攻击",def:"防御",spd:"速度"}).map(([stat,name]) => `<div><dt>${name}</dt><dd>${stats[stat].toLocaleString()}</dd></div>`).join("")}</dl><h4 class="at-subheading">随身装备</h4><div class="at-kit">${rules.slots.map(slot => {
    const quote = atelierGearQuote(id,slot.id);
    return `<button type="button" data-growth-gear="${slot.id}" aria-label="培养${esc(quote.name)}" title="${esc(quote.name)}" onclick="openGrowthGear('${id}','${slot.id}')">${atelierGearArt(quote)}<b>${esc(slot.name)}</b><small>+${quote.level} · ${quote.star}星</small></button>`;
  }).join("")}</div></section><section class="at-agenda"><small class="at-label">下一步培养</small><h4>${esc(goal.title)}</h4><p>${esc(goal.note)}</p><button type="button" class="ws-command" onclick="${goal.action === "cultivation" ? `openWorkshop('${id}','overview')` : "followGrowthGoal()"}">前往对应玩法 ›</button><div class="at-index"><span>星阶</span><b>${next ? `下一档 ${next.star} 星 · ${esc(next.name)}` : milestones.length ? "星阶效果已全部生效" : `${own.star}/${growthStarMax()} 星`}</b><span>收藏羁绊</span><b>${bonds.filter(bond => bond.percent).length}/${bonds.length} 条已生效</b></div><nav class="at-contents" aria-label="角色工坊">${options}</nav></section></div></article>`;
}
