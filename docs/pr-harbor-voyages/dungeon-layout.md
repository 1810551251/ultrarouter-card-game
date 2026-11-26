# 副本与船长入口 · 设计对照

参考范围是随需求提供的「关卡设计」与「主界面」前端源代码。游戏保留星霜的四章名称、敌人素材、模型数值与账号体系。海岸、路线、地标、天气和罗盘均由 SVG 绘制；没有用截图拼接地图，也没有导入参考工程的本地存档或奖励规则。

| 参考代码 | 星霜实现 | 对玩家的意义 |
| --- | --- | --- |
| `主界面/features.js` 的 `profile()` | 顶栏或首页头像 → `js/captain.js` 的 `openCaptainProfile()` | 在同一处查看昵称、ID、公会、等级/经验、战力、体力、整卡收藏率和副本星数，也可直接进入对应榜单。 |
| `关卡设计/index.html` 的 `.map-panel > .map-scroll > #seaMap` | `js/voyage.js` 的 `renderVoyage()`；`voyage-screen.css` | 全屏显示副本。左上返回和章节名称、右上钱包、左侧翻章、左下难度、右下宝箱独立浮在地图上。点击章节名称可切换副本分类。 |
| `关卡设计/voyage-map.js` 的 `positions()`、`routeSegment()`、`island()`、`oceanArt()` | `js/voyage-map.js`；`voyage-map.css` | 保留 820×510 横屏、400×640 竖屏、820×260 矮横屏三套坐标，以及海岸等深线、可航/已航路线、地标和头像航点。 |
| `关卡设计/app.js` 的 `showDetail()`、`showChest()` | `js/voyage.js` 的 `voyageDetail()`、`voyageChest()` | 进入关卡前可对照三星条件、敌方阵容、消耗和奖励。三档星箱各自显示已领取、可领取或还差几星；首页星箱入口优先定位已有奖励。 |
| `关卡设计/map-screen.css` 的 `.map-bottom-layer` 与移动端断点 | `voyage-screen.css` | 竖屏难度纵排，箱图置于星数上方；矮横屏压缩控制栏，为航点留出独立区域。 |

星数使用服务端保存的关卡历史最高星。首页和榜单只统计主线三难度，按关卡 ID 去重；资源副本与元素裂隙不计入。每章、每个难度的宝箱独立记录；查看宝箱和打开资料都不产生奖励操作。
