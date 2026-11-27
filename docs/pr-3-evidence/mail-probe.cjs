const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require('playwright-core');

const root = path.resolve(__dirname, '../..');
const origin = 'http://pr3-fixture.invalid';
const save = {mats:{field_notes:7, training_manual:11}, mailbox:{messages:[{
  id:'review-stage-reward', title:'副本结算 · 领航抄录所', sender:'复审测试夹具',
  body:'本邮件的附件和服务端回执由测试夹具给定。', created_at:1795753800000,
  resources:{field_notes:2, training_manual:1}, claimed:false, read:true
}]}};
const documentHtml = `<!doctype html><html lang="zh-CN"><meta charset="utf-8"><title>PR #3 邮箱交互验证</title>
<link rel="stylesheet" href="/harbor-mail.css"><style>*{box-sizing:border-box}body{background:#152635;color:#fff;font:18px system-ui;margin:2rem auto;max-width:1050px}header p{color:#f6d99f}button,input{font:inherit}.mail-page{position:static}main{padding:1rem}</style>
<body data-page="page-mail"><header><h1>邮箱领取 · 前端交互测试</h1><p>虚构存档和固定回执，仅验证界面，不连接游戏数据库。</p><b id="counts"></b></header><main class="mail-page"><div id="harbor-mail-body"></div></main>
<script src="/mail.js"></script><script>
const MATS={field_notes:{name:'斥候札记'},training_manual:{name:'历练手记'}};
const esc=value=>String(value).replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('"','&quot;');
let SAVE;const Store={ready:()=>!!SAVE,sessionEpoch:1,profile:{},mutate:async(action,fields)=>{
  const response=await fetch('/fixture/claim',{method:'POST',body:JSON.stringify({action,...fields})});
  const packet=await response.json();SAVE=packet.save;renderTopbar();return packet;
}};
function renderTopbar(){document.querySelector('#counts').textContent='样例背包：斥候札记 '+SAVE.mats.field_notes+'，历练手记 '+SAVE.mats.training_manual;}
fetch('/fixture/profile').then(response=>response.json()).then(state=>{SAVE=state;renderTopbar();renderHarborMail()});
</script></body></html>`;

async function run() {
  const browser = await chromium.launch({channel:'msedge', headless:true});
  try {
    const page = await browser.newPage({viewport:{width:1200,height:680}});
    const errors = [], requests = [], rows = [];
    let profiles = 0, release;
    const hold = new Promise(resolve => { release = resolve; });
    page.on('pageerror', error => errors.push(error.message));
    await page.route('**/*', async route => {
      const url = new URL(route.request().url());
      assert.equal(url.origin,origin);
      if (url.pathname === '/') return route.fulfill({contentType:'text/html; charset=utf-8',body:documentHtml});
      if (url.pathname === '/mail.js') return route.fulfill({contentType:'text/javascript',path:path.join(root,'js/mail.js')});
      if (url.pathname === '/harbor-mail.css') return route.fulfill({contentType:'text/css',path:path.join(root,'harbor-mail.css')});
      if (url.pathname === '/fixture/profile') { profiles++; return route.fulfill({json:save}); }
      if (url.pathname === '/fixture/claim') {
        requests.push(JSON.parse(route.request().postData()));
        await hold;
        save.mats = {field_notes:9, training_manual:12};
        save.mailbox.messages[0].claimed = true;
        return route.fulfill({json:{save,drops:['斥候札记 +2','历练手记 +1']}});
      }
      return route.abort();
    });
    async function record(checkpoint, expected, filename) {
      const actual = await page.evaluate(() => ({materials:[SAVE.mats.field_notes,SAVE.mats.training_manual],claimed:SAVE.mailbox.messages[0].claimed,disabled:document.querySelector('[data-mail-claim]').disabled}));
      assert.deepEqual({...actual,claimRequests:requests.length,profileReads:profiles},expected);
      rows.push({checkpoint,...expected});
      if (filename) await page.screenshot({path:path.join(__dirname,filename)});
    }
    const claim = page.locator('[data-mail-claim]');
    await page.goto(origin);
    await claim.waitFor();
    await record('before',{materials:[7,11],claimed:false,disabled:false,claimRequests:0,profileReads:1},'01-before.png');
    const firstRequest = page.waitForRequest('**/fixture/claim');
    await claim.dblclick();
    await firstRequest;
    await record('double-click, response held',{materials:[7,11],claimed:false,disabled:true,claimRequests:1,profileReads:1},null);
    assert.deepEqual(requests,[{action:'mail_claim',item_id:'review-stage-reward'}]);
    release();
    await page.waitForFunction(() => SAVE.mailbox.messages[0].claimed && document.querySelector('.mail-empty'));
    await page.locator('[data-mail-tab="all"]').click();
    await record('confirmed',{materials:[9,12],claimed:true,disabled:true,claimRequests:1,profileReads:1},null);
    await page.reload();
    await page.locator('[data-mail-tab="all"]').click();
    await claim.evaluate(button=>button.click());
    await record('reload, attempt to claim again',{materials:[9,12],claimed:true,disabled:true,claimRequests:1,profileReads:2},'02-reloaded.png');
    assert.deepEqual(errors,[]);
    const report={engine:browser.version(),database:null,fixture:true,materialOrder:['field_notes','training_manual'],rows,errors};
    fs.writeFileSync(path.join(__dirname,'browser-results.json'),JSON.stringify(report,null,2)+'\n');
    console.log(JSON.stringify(report));
  } finally { await browser.close(); }
}
run().catch(error=>{console.error(error.message);process.exitCode=1;});
