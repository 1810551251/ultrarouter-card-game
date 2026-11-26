(() => {
  'use strict';
  const esc = value => String(value ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  const themes = {
    dawn: { name: 'MIST HARBOR', current: '温暖洋流', land: '#3e6258', coast: '#7a9681', ridge: '#a3ad86', water: '#172e3b', light: '#294c53', ink: '#b4c9b9', landmark: 'harbor', points: [[87,345],[245,286],[391,361],[552,326],[711,226],[570,124],[373,161],[183,131]] },
    desert: { name: 'EMBER DOCK', current: '干燥季风', land: '#847452', coast: '#bfa16d', ridge: '#e4c88d', water: '#26383f', light: '#425858', ink: '#d6c49d', landmark: 'palace', points: [[83,352],[233,277],[387,354],[566,337],[719,231],[557,145],[379,145],[195,126]] },
    sky: { name: 'MOON TIDE', current: '上升气流', land: '#63858a', coast: '#a9c7c3', ridge: '#d6dfc3', water: '#203745', light: '#3b626d', ink: '#c4dce0', landmark: 'bell', points: [[91,351],[238,278],[393,355],[560,315],[712,218],[568,131],[373,166],[187,127]] },
    justice: { name: 'STAR MERIDIAN', current: '漩涡海流', land: '#536b70', coast: '#91a3a0', ridge: '#c1c4b4', water: '#1c2b3b', light: '#334c5d', ink: '#bacbd1', landmark: 'gate', points: [[85,351],[236,277],[384,355],[559,327],[714,222],[570,126],[373,160],[188,129]] }
  };
  const lockIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><rect x="6" y="10" width="12" height="10" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3m-4 4v2"/></svg>';
  const crownIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true"><path d="m3 7 5 5 4-8 4 8 5-5-3 12H6Z"/><path d="M7 16h10"/></svg>';
  const layouts = {
    landscape: { width:820, height:510 },
    portrait: { width:400, height:640 },
    compact: { width:820, height:260 }
  };
  function detectLayout() {
    const width = globalThis.innerWidth || 1440, height = globalThis.innerHeight || 900;
    return width < 700 && height > width ? 'portrait' : height < 600 ? 'compact' : 'landscape';
  }
  function positions(theme, count, layout) {
    const portrait = layout === 'portrait', compact = layout === 'compact';
    if (count === 2) return portrait ? [[103,505],[294,133]] : compact ? [[228,188],[598,63]] : [[215,362],[618,136]];
    if (count === 3) return portrait ? [[105,537],[293,320],[105,100]] : compact ? [[109,188],[413,62],[712,185]] : [[118,369],[413,113],[702,345]];
    if (count === 1) return [[layouts[layout].width / 2, layouts[layout].height / 2]];
    const route = portrait ? [[95,530],[305,530],[302,376],[97,376],[95,223],[306,223],[300,70],[99,70]]
      : compact ? [[95,181],[302,177],[516,183],[725,177],[725,55],[514,54],[302,58],[95,53]]
      : theme.points.map(([x,y]) => [x, (y - 124) * 1.28 + 94]);
    return Array.from({ length: count }, (_, i) => {
      const t = i * (route.length - 1) / (count - 1), a = Math.floor(t), b = Math.min(a + 1, route.length - 1), f = t - a;
      return [route[a][0] * (1 - f) + route[b][0] * f, route[a][1] * (1 - f) + route[b][1] * f];
    });
  }
  function routeSegment(points, i) {
    const a = points[i], b = points[i + 1], previous = points[Math.max(0, i - 1)], next = points[Math.min(points.length - 1, i + 2)];
    const curve = points.length < 4 ? .28 : .18;
    let c1 = [a[0] + (b[0] - previous[0]) * curve, a[1] + (b[1] - previous[1]) * curve];
    let c2 = [b[0] - (next[0] - a[0]) * curve, b[1] - (next[1] - a[1]) * curve];
    if (points.length === 2) { const dx = b[0] - a[0], dy = b[1] - a[1]; c1 = [a[0] + dx * .8, a[1] + dy * .12]; c2 = [b[0] - dx * .8, b[1] - dy * .12]; }
    return `M${a.join(' ')} C${c1.join(' ')} ${c2.join(' ')} ${b.join(' ')}`;
  }
  function island(x, y, index, large, theme, layout) {
    const portrait = layout === 'portrait', compact = layout === 'compact';
    const rx = portrait ? (large ? 69 : 56 + index % 2 * 5) : large ? 87 : 58 + index % 3 * 7;
    const ry = compact ? (large ? 33 : 25 + index % 2 * 4) : portrait ? (large ? 42 : 35 + index % 2 * 4) : large ? 44 : 29 + index % 2 * 7;
    const points = Array.from({ length: 16 }, (_, i) => {
      const a = Math.PI * 2 * i / 16, irregular = 1 + .17 * Math.sin(i * 2.8 + index * 2.4);
      return [Math.cos(a) * rx * irregular, Math.sin(a) * ry * irregular];
    });
    let d = `M${((points[15][0] + points[0][0]) / 2).toFixed(1)} ${((points[15][1] + points[0][1]) / 2).toFixed(1)}`;
    points.forEach((p, i) => { const next = points[(i + 1) % points.length]; d += `Q${p[0].toFixed(1)} ${p[1].toFixed(1)} ${((p[0] + next[0]) / 2).toFixed(1)} ${((p[1] + next[1]) / 2).toFixed(1)}`; });
    d += 'Z';
    const foliage = theme.landmark === 'palace'
      ? `<path d="m-40 9 14-24L-4 9m17-3 15-24L48 6" fill="${theme.ridge}" opacity=".28"/><path d="m-26-15 2 24H-4m32-27 1 24h19" fill="${theme.coast}" opacity=".45"/>`
      : theme.landmark === 'gate'
        ? `<path d="M-46 2v-16h15V2m43-4v-22h16V-2m8 6v-13h13V4" fill="${theme.ridge}" opacity=".24"/><path d="M-42-10h7m-7 5h7m54-10h6m-6 6h6" stroke="${theme.ridge}" opacity=".3"/>`
        : `<g fill="${theme.ridge}" opacity=".45"><path d="m-42 8 8-19L-25 8h-6v6h-6V8Zm52-2 9-22L29 6h-7v6h-6V6Zm23 9 7-17 8 17h-5v5h-5v-5Z"/></g>`;
    return `<g class="voyage-island" transform="translate(${x} ${y + 10})"><path d="${d}" fill="none" stroke="${theme.coast}" stroke-width="22" opacity=".06"/><path d="${d}" fill="none" stroke="${theme.coast}" stroke-width="11" opacity=".13"/><path d="${d}" transform="translate(0 6)" fill="#091924" opacity=".7"/><path d="${d}" fill="${theme.land}" stroke="${theme.coast}" stroke-width="1.6"/><path d="${d}" fill="none" stroke="${theme.ridge}" stroke-width=".8" transform="scale(.82 .76)" opacity=".28"/>${foliage}</g>`;
  }
  function landmark(theme, count, layout) {
    const [x, y] = layout === 'portrait' ? [201,322] : layout === 'compact' ? [411,130] : count < 4 ? [415,355] : [458,239];
    const art = {
      harbor: '<path d="M-31 17H32M-21 10h42M-12 10v-22H6v22M-15-12H9L3-20H-9ZM-6-20v-14m0 6h13l-4 6H-6"/><path d="M-8-8H2v8H-8Z" fill="currentColor" opacity=".45"/><path d="m-28 16 3-10m47 10L19 6"/>',
      palace: '<path d="M-36 16h72M-27 16V-8h14v24m26 0V-8h14v24M-30-8l10-10L-10-8Zm40 0 10-10L30-8ZM-11 16V-6h22v22M-14-6Q0-34 14-6ZM-4 16V5a4 4 0 0 1 8 0v11M0-26v-10"/>',
      bell: '<path d="M-30 17h60M-20 16v-38h40v38M-24-22h48l-9-8h-30ZM-13 4h26L9-4v-8a9 9 0 0 0-18 0v8Zm9 4a4 4 0 0 0 8 0M0-22v-5"/><path d="M-30-3q-5-8 0-15m60 15q5-8 0-15" stroke-dasharray="2 3"/>',
      gate: '<path d="M-37 17h74M-29 17v-35h12v35M17 17v-35h12v35M-33-18h20v-7h-20Zm46 0h20v-7H13ZM-17 17V-7Q0-28 17-7v24M-12 17V-5Q0-19 12-5v22M0-12v29"/><path d="M-25-12h4m0 8h-4m46-8h4m-4 8h4"/>'
    }[theme.landmark];
    return `<g class="voyage-landmark" transform="translate(${x} ${y})" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round">${art}<text y="37" stroke="none" fill="currentColor" text-anchor="middle">${esc(theme.caption)}</text></g>`;
  }
  function oceanArt(theme, chapterId) {
    const prefix = `voyage-${chapterId}`;
    return `<svg class="voyage-ocean" viewBox="0 0 820 510" preserveAspectRatio="none" aria-hidden="true">
      <defs><radialGradient id="${prefix}-water" cx="45%" cy="42%" r="75%"><stop stop-color="${theme.light}"/><stop offset="1" stop-color="${theme.water}"/></radialGradient><pattern id="${prefix}-grid" width="52" height="51" patternUnits="userSpaceOnUse"><path d="M52 0H0V51" fill="none" stroke="#91b3b8" stroke-width=".65" opacity=".095"/></pattern><linearGradient id="${prefix}-mist"><stop stop-color="#a6cbd0" stop-opacity="0"/><stop offset=".5" stop-color="#a6cbd0" stop-opacity=".06"/><stop offset="1" stop-color="#a6cbd0" stop-opacity="0"/></linearGradient></defs>
      <rect width="820" height="510" fill="url(#${prefix}-water)"/><rect width="820" height="510" fill="url(#${prefix}-grid)"/>
      <g class="voyage-depth" fill="none" stroke="${theme.ink}" stroke-width=".7" opacity=".085"><path d="M-40 217C95 80 246 445 422 276S721 4 890 180"/><path d="M-30 239C113 110 236 470 438 305S739 34 887 204"/><path d="M-20 260C104 139 242 487 452 329S754 61 885 229"/><path d="M-10 284C98 166 246 510 471 351S776 84 890 250"/><path d="M-50 350C95 223 224 555 471 416S771 153 890 312"/></g>
      <g class="voyage-waves" fill="none" stroke="${theme.ink}" stroke-linecap="round" stroke-width="1.1" opacity=".2"><path d="M80 207q7-5 14 0t14 0m-21 7q7-4 14 0M315 267q7-5 14 0t14 0m-21 7q7-4 14 0M667 402q7-5 14 0t14 0m-21 7q7-4 14 0M675 80q7-5 14 0t14 0M42 119q7-5 14 0t14 0M568 425q7-5 14 0t14 0"/></g>
      <g fill="${theme.coast}" opacity=".25"><path d="m57 265 5-6 13 2 3 6-12 4Zm675 95 4-8 11 4 3 7-12 4ZM303 74l8-6 8 4-2 7-12 1ZM48 432l3-7 7 1 4 6-8 4Z"/></g>
      ${theme.landmark === 'bell' ? '<g class="voyage-clouds" fill="#c1d9dc" opacity=".065"><path d="M48 226q-12-28 16-28 5-27 31-20 19-15 34 11 33 0 24 27Z"/><path d="M522 84q-12-23 13-26 4-20 27-16 15-13 28 8 26 0 20 23Z"/><path d="M591 396q-10-22 14-26 4-20 27-16 15-13 28 8 27 0 20 23Z"/></g>' : ''}
    </svg>`;
  }
  function mapArt(theme, points, stages, getProgress, getGate, selectedId, layout) {
    const { width, height } = layouts[layout];
    const routes = points.slice(0, -1).map((_, i) => {
      const target = stages[i + 1], completed = getProgress(target.id).stars > 0, available = getGate(target.id).open;
      const state = completed ? 'sailed' : available ? 'next' : 'unexplored';
      return `<g class="voyage-route ${state}${target.id === selectedId ? ' route-focus' : ''}"><path class="route-halo" d="${routeSegment(points, i)}"/><path class="route-line" d="${routeSegment(points, i)}"/>${available || completed ? `<path class="route-travel" d="${routeSegment(points, i)}"/>` : ''}</g>`;
    }).join('');
    const shipX = layout === 'portrait' ? 200 : layout === 'compact' ? 620 : 300, shipY = layout === 'portrait' ? 470 : layout === 'compact' ? 132 : 455;
    return `<svg class="voyage-chart" viewBox="0 0 ${width} ${height}" preserveAspectRatio="none" aria-hidden="true">${points.map(([x,y], i) => island(x, y, i, i === stages.length - 1, theme, layout)).join('')}${routes}${landmark(theme, stages.length, layout)}<g class="voyage-ship" transform="translate(${shipX} ${shipY})" stroke="${theme.ink}" stroke-linejoin="round" stroke-width="1.2"><path d="m-15 3 7 9H9l8-9Z" fill="${theme.coast}" fill-opacity=".45"/><path d="M0-23V3m-2-23L-15 0H-2m6-17L14 0H4" fill="${theme.ink}" fill-opacity=".65"/><path d="M-22 16h40m-31 4H9" opacity=".3"/></g></svg>`;
  }
  function compass(theme) {
    return `<svg class="voyage-compass" viewBox="-42 -50 84 92" fill="none" stroke="${theme.ink}" stroke-width=".8" aria-hidden="true"><circle r="29" opacity=".28"/><circle r="23" opacity=".18"/><path d="M0-35V35M-35 0H35m-21-14 28 28m0-28L-14 14" opacity=".35"/><path d="m0-23 5 18 18 5-18 5-5 18-5-18-18-5 18-5Z" opacity=".5"/><path d="M0-23 5-5 0 0-5-5Z" fill="${theme.ink}" opacity=".6"/><text x="0" y="-42" text-anchor="middle" fill="${theme.ink}" stroke="none" font-size="9">N</text></svg>`;
  }
  function render({ chapter, mode, stages, selectedId, getProgress = () => ({ stars: 0 }), getGate = () => ({ open: false }), layout = detectLayout() }) {
    if (!layouts[layout]) layout = detectLayout();
    const theme = {...(Object.values(themes)[chapter.region] || themes.dawn), caption:chapter.landmark}, points = positions(theme, stages.length, layout);
    const { width, height } = layouts[layout];
    const buttons = stages.map((s, i) => {
      const p = getProgress(s.id), gate = getGate(s.id), current = gate.open && !p.stars && selectedId === s.id, final = s.hasBoss && i === stages.length - 1, selected = selectedId === s.id;
      const status = p.stars ? '已通关' : gate.open ? '待挑战' : '未解锁';
      const marker = final ? `${crownIcon}区域首领` : current ? '<span class="voyage-live-dot"></span>下一站' : s.elite ? '精英遭遇' : '';
      return `<button type="button" class="map-node voyage-node${!gate.open ? ' locked is-locked' : ''}${p.stars ? ' complete is-complete' : ''}${current ? ' current is-current' : ''}${selected ? ' selected is-selected' : ''}${final ? ' is-boss' : ''}${s.elite ? ' elite is-elite' : ''}" data-voy-point="${esc(s.id)}" style="left:${points[i][0] / width * 100}%;top:${points[i][1] / height * 100}%" title="${esc(gate.reason || status)}" aria-label="${chapter.number}-${s.number} ${esc(s.name)}，${status}，${p.stars} 星${final ? '，区域首领' : ''}" aria-pressed="${selected}"><span class="voyage-node-marker">${marker}</span><span class="voyage-node-portrait"><img src="${esc(s.portrait)}" alt="" loading="eager" draggable="false">${!gate.open ? `<span class="voyage-node-lock">${lockIcon}</span>` : ''}<span class="voyage-node-index">${p.stars ? '✓' : String(s.number).padStart(2, '0')}</span></span><span class="voyage-node-label">${esc(s.name)}</span><span class="voyage-node-stars" aria-hidden="true">${s.material ? esc(s.material) : [1,2,3].map(n => `<span class="${p.stars >= n ? 'earned' : ''}">★</span>`).join('')}</span></button>`;
    }).join('');
    return `<div class="voyage-world" data-theme="${esc(chapter.id)}" data-mode="${esc(mode.id)}" data-layout="${layout}" style="--voyage-ink:${theme.ink};--voyage-coast:${theme.coast};--voyage-water:${theme.water}">${oceanArt(theme, chapter.region)}<div class="voyage-atmosphere" aria-hidden="true"></div><div class="voyage-route-plane">${mapArt(theme, points, stages, getProgress, getGate, selectedId, layout)}${buttons}</div>${compass(theme)}<div class="voyage-cartography" aria-hidden="true"><span>${theme.name}</span><small>${esc(chapter.name)}</small></div><span class="voyage-weather" aria-hidden="true"><span>↝</span> ${theme.current}</span></div>`;
  }
  globalThis.VoyageMap = Object.freeze({ render, detectLayout });
})();
