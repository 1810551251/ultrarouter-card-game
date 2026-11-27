import asyncio
import json
import os
import sys
import time
import uuid
from pathlib import Path

from playwright.async_api import async_playwright


async def main():
    origin, letter = sys.argv[1:]
    baseline = json.loads(Path('/pr3-state/pr3-fixture.json').read_text())['Assets']
    output = Path('/out')
    errors, requests = [], []
    profiles = 0
    async with async_playwright() as driver:
        browser = await driver.chromium.launch(executable_path='/usr/bin/chromium', args=['--disable-dev-shm-usage'])
        page = await browser.new_page(viewport={'width': 1160, 'height': 820}, locale='zh-CN')
        page.on('pageerror', lambda exc: errors.append(str(exc)))
        bundle = {'access_token': os.environ['REVIEW_BROWSER_KEY'], 'token_type': 'Bearer', 'access_expires_at': int(time.time()) + 600,
                  'user': {'id': int(os.environ['REVIEW_USER'])}, 'session': {'sid': str(uuid.uuid4()), 'current': True}}
        release = asyncio.Event()
        response_ready = asyncio.Event()

        async def route_request(route):
            nonlocal profiles
            url = route.request.url
            if not url.startswith(origin + '/'):
                await route.abort()
            elif url.endswith('/api/user/auth/refresh'):
                await route.fulfill(json={'success': True, 'data': bundle})
            elif url.endswith('/api/user/gacha/action'):
                intent = route.request.post_data_json
                requests.append(intent)
                response = await route.fetch(max_retries=0)
                if len(requests) == 1:
                    response_ready.set()
                    await release.wait()
                await route.fulfill(response=response)
            else:
                if url.endswith('/api/user/gacha/profile'):
                    profiles += 1
                await route.continue_()

        await page.route('**/*', route_request)
        await page.goto(origin)
        await page.locator('#app:not(.hidden) #page-home.active').wait_for()
        await page.locator('.nav-btn[data-nav="page-squad"]').click()
        slots = page.locator('#squad-slots .team-slot.filled')
        slot_count = await slots.count()
        assert slot_count == len(baseline['Squad'])
        assert await page.evaluate('SAVE.squad') == baseline['Squad']
        for slot, card in enumerate(baseline['Squad']):
            star, level, *_ = baseline['Cards'][card]
            assert f'Lv{level} · {star}星' in await slots.nth(slot).inner_text()
        await page.screenshot(path=output/'03-upgraded.png')
        await page.locator('.nav-btn[data-nav="page-home"]').click()
        await page.locator('[data-home-mail]').click()
        button = page.locator(f'[data-mail-claim="{letter}"]')
        await button.wait_for()
        assert await button.is_enabled()
        await page.screenshot(path=output/'04-attachment.png')
        counter = '[SAVE.mats.field_notes || 0, SAVE.mats.training_manual || 0]'
        before = await page.evaluate(counter)
        await button.dblclick()
        await asyncio.wait_for(response_ready.wait(), timeout=10)
        assert len(requests) == 1 and requests[0]['item_id'] == letter
        assert await button.is_disabled() and await page.evaluate(counter) == before
        release.set()
        await page.wait_for_function('id => SAVE.mailbox.messages.some(m => m.id === id && m.claimed)', arg=letter)
        credited = await page.evaluate(counter)
        assert [b - a for a, b in zip(before, credited)] == [2, 1]
        reads_before_reload = profiles
        await page.reload()
        await page.locator('#app:not(.hidden)').wait_for()
        if not await page.locator('#page-mail.active').count():
            await page.locator('[data-home-mail]').click()
        await page.locator('[data-mail-tab="all"]').click()
        await button.wait_for()
        assert profiles > reads_before_reload and await button.is_disabled()
        await button.evaluate('(button) => button.click()')
        assert len(requests) == 1
        try_again = await page.evaluate('''async id => {
          try { await Store.api(SITE.ep.action, 'POST', {action:'mail_claim',item_id:id,request_id:crypto.randomUUID()}); return 'accepted'; }
          catch (error) { return error.code; }
        }''', letter)
        assert try_again == 'ALREADY_CLAIMED' and len(requests) == 2
        assert requests[1]['request_id'] != requests[0]['request_id']
        await page.evaluate('Store.refresh().then(() => renderHarborMail())')
        assert await page.evaluate(counter) == credited and await button.is_disabled()
        await page.screenshot(path=output/'05-claimed.png')
        assert not errors, errors
        report = {'browser':browser.version, 'squad_slots':slot_count, 'before':before, 'confirmed':credited,
                  'double_click_claim_requests':1, 'after_reload':await page.evaluate(counter), 'profiles':profiles,
                  'forced_second_request':try_again, 'page_errors':errors}
        (output/'browser.json').write_text(json.dumps(report, indent=2)+'\n')
        await browser.close()


asyncio.run(main())
