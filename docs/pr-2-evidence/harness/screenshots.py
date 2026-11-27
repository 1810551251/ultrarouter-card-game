"""Capture UI fixtures, without using the captures as backend evidence."""
from datetime import datetime, timedelta, timezone
from functools import partial
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from threading import Thread
import json
import subprocess
import sys
import time
from urllib.parse import urlsplit

from playwright.sync_api import expect, sync_playwright


root = Path(__file__).resolve().parents[3]
out = Path(sys.argv[1])
out.mkdir(parents=True, exist_ok=True)
edge = sys.argv[2] if len(sys.argv) > 2 else None
day = datetime.now(timezone(timedelta(hours=8))).date()
stamp = int(time.time() * 1000)


class QuietFiles(SimpleHTTPRequestHandler):
    def log_message(self, *_):
        pass


def catalog(name):
    return json.loads((root / f'backend/model/gacha_{name}.json').read_text(encoding='utf-8'))


def capture(browser, origin, viewport, mobile):
    errors, shots, actions = [], [], []
    fail_profile = False
    revision = 1
    saved_rank, attempts = 101, 5
    dispatch = catalog('dispatch')
    days = [{'date': str(date), 'activities': ['training', 'wish'], 'training_left': 0, 'wish_targets': [],
             'marks_claimed': False, 'permits_claimed': False,
             'rules': {'wishes_per_day': 3, 'marks_per_day': 40, 'permits_per_day': 5}}
            for date in [day, day - timedelta(days=1)]]
    save = {'squad': ['c01', 'c04'], 'girls': {card: {'level': 24, 'star': 3} for card in ['c01', 'c04']},
            'gold': 1234, 'stamina': 90, 'staminaMax': 120, 'staminaTs': stamp, 'power': 7210,
            'mats': {'field_notes': 7, 'gear_plate': 4}, 'cleared': {}, 'stars': {},
            'progression': {'days': days, 'weeks': [], 'permits': 5, 'marks': 80, 'starter_claimed': True, 'trial_wins': {}, 'dual': {}},
            'daily': {'date': str(day), 'dispatch': {'contract': {'id': 'courier', 'leader': 'c01', 'progress': 2, 'claimed': False}}},
            'dispatch': {'relics': {}, 'mosaic': {'week': '2026-W48', 'cleared': 0, 'marks': [0, 4, 8]}}}
    squad = [{'id': 'c01', 'level': 24, 'star': 3}, {'id': 'c04', 'level': 24, 'star': 3}]
    opponents = [{'id': -rank, 'rank': rank, 'name': f'星霜守卫 {rank:03d}', 'power': 2600+15*rank, 'squad': squad[:1]}
                 for rank in [48, 60, 72, 84, 100]]
    materials = {'mat_r': 40, 'mat_sr': 20, 'mat_ssr': 5}
    history = []
    account = {'user_id': 7, 'name': '隔离测试旅人', 'quota': 2500000, 'today': {}, 'save': save,
               'progression': {'rules': {'awakening_costs': [80, 120, 200]}},
               'dispatch': dispatch, 'items': catalog('items'), 'cases': catalog('cases'),
               'cultivation': catalog('cultivation'), 'equipment': catalog('equipment')}
    page = browser.new_page(viewport=viewport, locale='zh-CN', reduced_motion='reduce', service_workers='block')
    page.on('pageerror', lambda error: errors.append(str(error)))

    def intercept(route):
        nonlocal fail_profile, revision, saved_rank, attempts
        url = route.request.url
        endpoint = urlsplit(url).path
        if not url.startswith(origin + '/'):
            return route.abort()
        if not endpoint.startswith('/api/'):
            return route.continue_()
        if endpoint.endswith('/auth/refresh'):
            data = {'access_token': 'synthetic-browser-fixture', 'token_type': 'Bearer', 'access_expires_at': int(time.time())+600,
                    'user': {'id': 7}, 'session': {'sid': 'synthetic-local-session', 'current': True}}
        elif endpoint.endswith('/profile'):
            if fail_profile:
                fail_profile = False
                return route.fulfill(status=429, json={'success': False, 'code': 'RATE_LIMITED'})
            data = {**account, 'revision': revision}
        elif endpoint.endswith('/ladder'):
            me = {'id': 7, 'rank': saved_rank, 'best_rank': saved_rank, 'name': account['name'],
                  'power': save['power'], 'squad': squad, 'cooldown_until': 0} if saved_rank else None
            data = {'me': me, 'remaining': attempts, 'limit': 5, 'total': 100+bool(me), 'server_now': stamp,
                    'next_settlement': stamp+3600000, 'next_reset': stamp+7200000, 'opponents': opponents, 'board': opponents,
                    'history': history, 'rewards': [{'max_rank': 100, 'materials': materials},
                    {'max_rank': 0, 'materials': {'mat_r': 30, 'mat_sr': 15, 'mat_ssr': 3}}],
                    'mail': [{'day': days[1]['date'], 'rank': 100, 'materials': materials, 'claimed': False}]}
        elif endpoint.endswith('/action'):
            intent = route.request.post_data_json
            actions.append(intent['action'])
            revision += 1
            data = {'save': save, 'revision': revision}
            if intent['action'] == 'ladder_challenge':
                assert set(intent) == {'action', 'request_id', 'count', 'opponent_id'}
                before, saved_rank = saved_rank, intent['count']
                attempts -= 1
                opponent = next(row for row in opponents if row['id'] == intent['opponent_id'])
                report = {'id': intent['request_id'], 'ladder': True, 'name': opponent['name'], 'win': True,
                          'own_before': before, 'own_after': saved_rank, 'attacker_name': account['name'], 'defender_name': opponent['name']}
                data['ladder'] = {**report, 'report': {'timed_out': False, 'duration_ms': 20, 'events': []}}
                history.append({**report, 'attacker_id': 7, 'defender_id': opponent['id'], 'attack_win': True,
                                'attacker_before': before, 'attacker_after': saved_rank, 'played_at': stamp})
            elif intent['action'] == 'ladder_leave':
                saved_rank = None
            elif intent['action'] == 'daily_contract_claim':
                date = next(row for row in days if row['date'] == intent['item_id'])
                date['marks_claimed'] = True
            elif intent['action'] == 'dispatch_contract_claim':
                assert intent['item_id'] == str(day)
                save['daily']['dispatch']['contract']['claimed'] = True
            else:
                raise AssertionError(intent['action'])
        else:
            data = {}
        route.fulfill(json={'success': True, 'data': data})

    def shot(name, locator):
        expect(locator).to_be_visible()
        locator.scroll_into_view_if_needed()
        page.evaluate('document.fonts.ready.then(() => true)')
        geometry = page.evaluate('({scroll: document.documentElement.scrollWidth, viewport: innerWidth})')
        assert geometry['scroll'] <= geometry['viewport'] + 1, geometry
        target = out / f'{name}.png'
        locator.screenshot(path=target, animations='disabled', timeout=12000)
        shots.append({'file': target.name, 'bytes': target.stat().st_size, 'layout': geometry})

    page.route('**/*', intercept)
    page.goto(origin + '/index.html')
    expect(page.locator('#app')).to_be_visible()
    page.locator('#page-home [data-nav="page-arena"]:not([data-arena-mode])').click()
    expect(page.locator('.ladder-opponent')).to_have_count(5)
    if mobile:
        shot('05-ladder-mobile', page.locator('.ladder-opponent').first)
    else:
        shot('01-ranked-challenges', page.locator('.arena-shell'))
        fail_profile = True
        page.locator('.arena-section-heading button').first.click()
        retry = page.get_by_role('button', name='重新读取竞技场')
        shot('02-transient-retry', retry.locator('..'))
        retry.click()
        page.locator('[data-ladder-challenge]').first.click()
        expect(page.locator('#modal-box')).to_contain_text('不消耗 E 余额')
        page.get_by_role('button', name='确认挑战 · 1 次').click()
        expect(page.locator('.arena-award')).to_contain_text('#101 → #48')
        shot('03-confirmed-rank-change', page.locator('#modal-box'))
        page.get_by_role('button', name='返回排名竞技').click()
        page.locator('[data-arena-tab="ladder-rewards"]').click()
        shot('04-daily-materials', page.locator('.ladder-mail'))
        page.locator('[data-arena-tab="ladder"]').click()
        page.get_by_role('button', name='退出排名竞技').click()
        shot('06-withdrawal-consent', page.locator('#modal-box'))
        page.get_by_role('button', name='确认退榜').click()
        expect(page.locator('.ladder-welcome')).to_be_visible()
        page.evaluate("nav('page-growth')")
        daily = page.locator('.growth-bank')
        expect(daily).to_be_visible()
        shot('07-today-and-banked-contracts', daily)
        daily.get_by_role('button', name='领40契约印').first.click()
        expect(daily.get_by_role('button', name='契约已寄')).to_be_disabled()
        page.evaluate("openWorkshop('c01', 'dispatch'); setDispatchRoom('contract')")
        claim = page.locator('[data-desk-action="dispatch_contract_claim"]')
        expect(claim).to_be_enabled()
        shot('08-independent-dispatch-claim', page.locator('#dispatch-panel'))
        claim.click()
        expect(claim).to_be_disabled()
        shot('09-dispatch-claimed', page.locator('#dispatch-panel'))
        assert actions == ['ladder_challenge', 'ladder_leave', 'daily_contract_claim', 'dispatch_contract_claim'], actions
    assert not errors, errors
    page.close()
    return {'viewport': viewport, 'captures': shots, 'actions': actions, 'page_errors': errors}


with ThreadingHTTPServer(('127.0.0.1', 0), partial(QuietFiles, directory=root)) as server:
    thread = Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        origin = f'http://127.0.0.1:{server.server_port}'
        with sync_playwright() as driver:
            browser = driver.chromium.launch(executable_path=edge)
            results = {'run_type': 'actual game UI with synthetic API fixtures', 'browser': browser.version,
                       'revision': subprocess.check_output(['git', '-C', str(root), 'rev-parse', 'HEAD'], text=True).strip(),
                       'runs': [capture(browser, origin, {'width': 1366, 'height': 900}, False),
                                capture(browser, origin, {'width': 390, 'height': 844}, True)]}
            browser.close()
        (out / 'browser.json').write_text(json.dumps(results, indent=2) + '\n', encoding='utf-8')
    finally:
        server.shutdown()
