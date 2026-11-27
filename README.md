# Starfrost Card Game

This repository contains **only the card game**, not the maintainer's website or the complete new-api platform.

Included: the browser game, 28 character definitions, game artwork, server-owned game rules, dungeon battles, progression, inventory, token shop, rankings, voluntary PvP matchmaking, custom rooms, and game regression tests.

## 星霜港航务

主页头像旁直接展示**船长等级与经验条**，点击头像可查看个人资料。船长经验仅来自**实际消耗体力（1 点体力 = 1 经验）**和**每日任务奖励**。卡牌不能超过船长等级，最高 150 级；老档承接已拥有伙伴的历史最高等级。每日点名、航战、训练、许愿与培养各有一项目标，完成后奖励自动寄入邮箱。北京时间 **09:00–12:00、12:00–21:00、21:00–24:00** 各能申请一次 30 体力补给邮件，超过自然恢复上限的补给会保留。

- **主线副本**采用全屏地图、沿途关卡、顶部资源栏、章节切换和底部难度与星箱布局。共四章 52 关，每章剧情 8 关、困难 3 关、噩梦 2 关，各自记录最高星、首通和首次三星；每章每种难度有三档一次性星箱，共 36 档。困难和噩梦分别从船长 8、15 级起，还要满足副本通关条件。
- **邮箱与兑换码**支持奖励附件、单封及一键领取、已领邮件整理。每日任务、定时补给、副本物资、星箱、活动与公会奖励寄入邮箱后领取。公开礼包码 `STARFROST2026` 与 `VOYAGEGIFT` 每账号每种限兑一次，有效期为北京时间 2026-11-27 至 2027-12-31。
- **资源航线**每天可分别挑战三次。领航抄录所产书卷，浮船锻造厂产装备工料，潮墨档案库产研训墨与晋阶章，星砂护商线产金币与徽记，深海星锻井产矿料与核心，棱镜打捞湾产宝芯券。达到三星的剧情、困难及资源航线可扫荡固定物资；不重复发首通、三星或站点余额。
- **港口活动**每周有 14 场，每天两场，每场最多三次提交。轮换玩法是装船配载、海雾定位、灯阵调配和潮汐择时；完成一场盖一枚章，3、7、11 枚可领取周奖励。活动使用游戏日与服务器题面，不接受客户端分数。
- **公会与排行**提供可翻页的招募列表，以及战力、收藏率、公会周航绩、竞技胜场、副本星级五种榜单。副本星级取四章三档难度每关的历史最高星，共 156 星。可点开榜内队伍；未登录也可在 `#rank-demo-stars` 查看独立标注的演示榜。会长可自选是否公开，成员每日可选一项港务共建，每周三线共同达到里程后由参与者领取奖励。

Git 仓库的[本次更新公告](docs/pr-harbor-voyages/README.md)附有桌面与手机版界面截图和对应说明，不随游戏源文件压缩包分发。

## 卡牌培养

打开「成长」或从「编队」点卡进入工坊：用顶部下拉框切换已获得的角色，卡面保留当前突破色、星级和实际属性，其余空间用于培养。培养和收藏羁绊不要求上阵；装备套装、战斗见证和队阵只看对应的实装、出战成员。

| 玩法 | 生效条件与用途 |
| --- | --- |
| 图鉴与收集羁绊 | 首次获得整卡即为全员增加固定属性；碎片不算。38 组收藏羁绊中 **36 组三人、2 组双人**，R / SR / SSR 每卡各有 2 / 4 / 6 条。集齐即为组员加成，不用上阵；全组均达 6 星后进阶。同一角色同属性的羁绊加算，最高 +30%。 |
| 角色升星与潜能 | 上限 **12 星**。升星需要**该角色专属碎片**和对应品质的结晶；6、7、8–12 星分别要求觉醒 1、2、3 阶。星阶路线逐项列出 3、6、9、12 星的具名战斗效果。例如单攻 9 星命中后按实际生命伤害吸血 5%，12 星每次普攻多回 10 点怒气；守护 12 星开战额外获得最大生命 12% 的护盾。潜能盘则在偶数星依次开放，每节点生命、攻击、防御各 +1%。 |
| 六维洗练 | 「潜能 → 个体洗练」可重掷生命、攻击、防御、速度、暴击和暴伤（每项 0–31）。看中哪一项就单独锁住；下次只洗未锁项，可能升也可能降。每次消耗 2 星锻矿，每锁一项多用 1 矿，六项全锁时不能洗练。锁定和解锁本身免费。 |
| 五部位装备 | 专武、头、手、身、足生来都是白装。按已有材料可自动强化本部位，+3 / +6 / +10 / +15 / +22 / +30 依次变绿、蓝、紫、金、橙、红。装备升星是独立路线：每次升一星，不能靠换色跳星；12 星开冠，之后还有 13–15 星超限。 |
| 卡牌突破 | 逐层检查当前卡级并扣料，不看星级。前三次突破分别消耗 8、13、19 份研训墨，从第三次起另耗晋阶章；换色后价格继续递增。同色增加突破加号，本色练满后进入下一色 +0（例如青绿 +2 → 苍蓝 +0）。卡牌等级和当前经验始终保留。 |
| 角色升级 | 当前上限为**船长等级**，最终上限 **150 级**。「升级 → 一键升级」可选经验书和数量，达到本次上限即停止扣书。Lv.120→121 单级 49,235 经验；Lv.120→150 共 2,372,425 经验。 |
| 努力值 | 生命、攻击、防御、速度、抵抗五维共 **12 段**，单项累计上限从 160 经 1,400、12,500 到 33,000。每卡每日 10 练，第 1 段五项从零练满需 80 次，第 12 段从前一段起点练满需 285 次；至少分别跨 8 和 29 个游戏日。选单项可自动练至当天额度、材料或阶段上限。全满晋段，数值继承；面板单列换算后的实战加成。 |
| 技能 | 普攻、怒气技和被动的可学上限均为**角色当前等级**。全账号技能点容量 20，每 20 秒回 1 点；也可主动花 120 游戏金币买 1 点。单升或自动升只用现有点数及金币，不暗中买点。换色不改变已学或生效的技能等级。 |
| 携带道具 | 用行旅徽记与指定工料确定制作六种道具。每卡一件，更换免费退回旧件。勘星手记、守望誓约两套：**上阵的不同角色携带同系列不同物件**才计入 2 / 3 档，仓库、替补和重复道具不凑数。与纯收集羁绊不同。 |
| 专武觉醒与铭刻 | 专武每级主属性 +2%；+3 开放刃纹、壁纹、律纹三选一。+5、+10 且角色觉醒达 1、3 阶，可消耗觉醒棱芯觉醒专武。每阶生命 +3%、开场怒气 +10。专武觉醒与装备升星是两项独立培养。 |
| 材料与回收 | 历练手记、斥候札记、远征图志、星轨秘典每本分别提供 600 / 1,200 / 3,600 / 7,200 经验。背包列出七色品质、配方和来源。多余专属碎片只换锻矿（R 每 10 枚、SR 5 枚、SSR 2 枚换 1 矿）；自动保留解锁或下一星用量和零头。旧尘只能每 10 枚同种换 1 矿，不可垫升星。 |

例如，「炉边誓约」需要灯里、剑华、龙音。只取得前两人时没有羁绊攻击；获得第三张整卡，三人立即各得攻击 +4%，三人都到 6 星提升为 +7%。竞技防守仍以最后保存的阵容快照为准。

物品用白、绿、蓝、紫、金、橙、红七档品质，色框之外还有品质名和纹章；五类宝石的切面、三系芯片的线路也不同。背包、制作配方、关卡掉落和结算共用物品图鉴。研训墨用于日常培养，晋阶章用于开新色和努力阶段，行旅徽记可指定制作携带物。

**宝石和芯片分三条轴。** 品质从白到红，决定基础强度与高品质专有效果；等级 **1–120** 主要提升基础属性；阶 **1–12** 主要增强触发效果。换卡、卸下不会重置。宝石看同族而非品质颜色，九孔盘有 3/6/9 件效果；芯片有定位的六槽盘及 2/4/6 件效果。可混搭，芯片锁位防误拆。

「召唤」的宝石池与芯片池独立单抽或十连。先取模板（五种宝石各 1/5；芯片三系六编号各 1/18），再独立确定品质。白 / 绿 / 蓝 / 紫 / 金 / 橙 / 红分别为 42% / 28% / 16.5% / 8.6% / 3.5% / 1.2% / 0.2%。每抽消耗一张对应游戏券；主线首领、裂隙分别掉落两类券，也可用 8 星锻矿兑换。

单件绑定装备从 +0 到 +30 合计 1,835 份对应工料；从 0 到 12 星另需 530 枚辉星铆钉。装备超限到 15 星、角色 12 星和宝芯 12 阶分别结算，没有互相替代的通用升星奖励。

## 夜巡与异闻

- **签下哪份誓约？** 灯使靠主线胜利留下见证，聆潮走裂隙，守阈要无减员归来。各卡独立，每日最多 3 印。印记累计至 3 / 9 / 21 并完成仪式才能晋阶；第二阶从两种分支择一，六种分支改变真实行动，例如「余烬标灯」首次施放怒气技后救援生命比例最低的队友，「逆潮」则抽走场上敌人的怒气。换分支要重新交证物与晋阶章。
- **案件怎么结？** 在异闻调查局选案和领队，逐幕取证、隐行或冒险，决定何时归档、追查或撤离。证物用于定相、晋阶与镇定，不是角色经验。账号每天最多开两件不同的案，证物有独立日上限；当前案件由服务端记录，刷新不会重开。
- **要不要借夜？** 借夜让誓约效果额外 +5，但开战先付最大生命的 12%；带卡取胜失序 +2，累积至 6 强制熄灯。守常胜利可慢慢恢复，或用回声类证物镇定。编队凑齐三条已晋阶誓约有开场回怒；同路不同分支能为该路成员提供开场护盾。收藏未上阵的卡不能凑队阵。

### 密务厅：七张能玩的桌子

| 桌位 | 玩家做什么 | 取舍与结算 |
| --- | --- | --- |
| 封存柜 | 从特定案件的特定结局取得制作资格，再用证物封存三件遗物；只选一件给编队第一位。 | 倒鸣铃换开场怒气但扣血，雨衣给盾但减普攻回怒，十三面镜增强怒技但削弱普攻。可免费归柜，不动旧战报。 |
| 雾钟校时 | 依照齿印的相对位置、首末位排除线索，把灰烬、弦月、潮汐排成唯一顺序。 | 每个游戏日两次机会；重排草稿不耗次，提交错误排列才消耗一次。 |
| 限期委托 | 选好领队后，在两人护送、速破裂隙、三色多关巡灯之间只签一份。 | 当日只认签约后开局的权威战报；先解读条款再组队，达标后主动交付，零点未交作废。 |
| 雾路运灯 | 规划四层三线驿路，左右每次最多挪一线。算清开支、补给和沿路货物。 | 第三层之前也可返程保货；未抵终点先耗尽灯油则失货，全程送达另得徽记。 |
| 星图拼版 | 把九格切成三个连通星座，每座一个星核；恰好一条直线、两个折角。 | 一周三幅，每幅可任意试摆，服务端验拼；不比手速，同色既有文字标记也有颜色。 |
| 镜厅试阵 | 从每日六张标准借卡中挑两张，决定先锋和配令，挑战守护、速攻或施技目标。 | 三试之内根据服务器战记调整；同日同阵同令可复现。不消耗自有体力，也不会强化或替换个人阵容。 |
| 暗语译台 | 猜四位六符密本；每次根据“同符同位”“同符异位”数量排除候选，重复符号不会重复计数。 | 北京日有奖密本限五猜，结束后仍可开无奖练习；答案不随存档发给客户端。 |

入门建议：先在「异闻调查」安全归档一次，把证物存下；再去密务厅校钟或试阵，它们不需要养好角色。已有队伍可接护送，在同一场主线胜利中兼做本卡的夜巡见证和账号委托；两条进度各验自己的条件。

Not included: the website dashboard, account-management implementation, API-provider relay, QQ bot/binding system, payment platform, production deployment scripts, server addresses, credentials, real accounts, or database files.

## 排名竞技场

排名挑战免费且不碰站点余额。报名保存当前防守阵容，允许其他玩家离线挑战；退出后立即退榜，已结算材料仍可领取。重新报名从榜尾开始，不重置次数或冷却。每天 5 次挑战，北京时间 05:00 重置，21:00 按结算时的名次发放材料。余额匹配和自定义大厅仍是独立入口，要求每场双方自愿且战力差不超过较低方的 20%。

[合并与复审证据](docs/pr-2-evidence/README.md)列出了游戏端的验证条件、结果与截图。宿主迁移和结算任务见 [Backend Integration](backend/INTEGRATION.md)。

## What Can Run Locally

Use Node.js 24+ and Python 3.10+:

```sh
npm ci
npm test
npm run test:release
python scripts/preview.py
```

Open `http://localhost:8080/gacha/` for a **read-only frontend preview**. The preview does not log anyone in, create a wallet, or simulate payable rewards. Playing requires your own authenticated backend. Editing browser storage never creates valid game assets.

The Go files are a **game extension**, not a standalone server. They currently use the host's authenticated sessions, wallet transaction/cache helpers, and database connection. The host implementation is deliberately not bundled. See `backend/INTEGRATION.md` for the required integration points; copying the files into an arbitrary new-api release without adapting those dependencies is not a supported one-click installation.

## Source Layout

| Path | Game Content |
| --- | --- |
| `index.html`, `style.css`, `mobile.css`, `js/` | Game interface, server-state synchronization, battle playback |
| `assets/` | Character, enemy, background and UI artwork; AI provenance |
| `backend/model/gacha*.go` | Game state, rewards, battles, progression, guilds, matchmaking and settlement |
| `backend/model/gacha_catalog.json` | Authoritative character and stage configuration |
| `backend/model/gacha_cultivation.json` | Card-specific fellowships, star passives, shard recycling, weapons, potential, individual values and artifact sets |
| `backend/model/gacha_equipment.json`, `backend/model/gacha_items.json` | Five-slot gear costs, progression and resonance; item qualities, categories, uses and sources |
| `backend/model/gacha_training.json`, `backend/model/gacha_carry.json` | Breakthrough colours/sublevels, effort and skill costs; held-item recipes, passives and team sets |
| `backend/controller/gacha*.go` | Session-gated game endpoints and intent validation |
| `backend/integration/` | Game-only route and migration examples for your own host |
| `tests/` | Browser authority/readability checks, preview and release-boundary checks |
| `scripts/` | Read-only preview, source audit, packaging and shared CPU-budget tooling |

All browser endpoints are same-origin relative paths. This project does not need or provide access to the original maintainer's server. Contributors supply their own test accounts and databases.

## Development And Release

Gameplay, rewards, inventory, E balances and settlement remain server-authoritative. Keep the host session guard, transaction boundaries, escrow, cache consistency and idempotency when adapting the backend. See `CONTRIBUTING.md`.

```sh
python scripts/audit-release.py
python scripts/package-release.py --name starfrost-game
```

The packager includes only game source and checks the game-only backend allowlist. Local data, secrets, dependencies and Git history are excluded. Private deployment-marker checks are described in `SECURITY.md`. Publish only the reviewed source allowlist or ZIP, never an entire host working directory. A new public repository should start with fresh Git history so private deployment history cannot be included.

Publishing this source does not deploy or connect to any production service. Check `VALIDATION.md` for the checks actually performed and the external-host limitation. The GitHub Actions workflow checks game source and frontend behavior; it does not deploy the game or prove compatibility with an arbitrary host backend.

## Attribution And Artwork

The existing game backend is based on [new-api / QuantumNous](https://github.com/QuantumNous/new-api). Its original identifiers and license/notice files are retained; preserving attribution does not mean the complete platform is included. Code follows the bundled GNU AGPLv3 terms and notices.

The maintainer confirmed that all 28 character images are AI-generated. Artwork licensing is separate from code licensing; see `ASSETS.md` and `assets/PROVENANCE.json`. Do not assume an unconfirmed image license permits commercial redistribution.
