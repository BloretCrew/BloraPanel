# 图标候选第二轮

用户否定 B 的渐变并要求更多方向。新增九套完整14图标候选 D–L；每套包含实际桌面、启动器、14图标近景、20/32/48 CSS像素及深底对比。无渐变，线条类共用现有 Lucide 库，块面和几何类使用本地 SVG。桌面布局与默认图标不变，开发预览参数 `?icon-study=d` 至 `l` 选择候选。

总览：[九方向截图](../../web/ui-screenshots/icon-explorations-02/overview.png)。浏览器查看入口：开发服务器 `/icon-explorations.html`，单方向添加 `?variant=d` 至 `l`。

|方向|设计变化|近景 / 尺寸 / 深底|实际桌面|启动器|
|---|---|---|---|---|
|D 瓷白|双色 Lucide 细线、白色底板|[查看](../../web/ui-screenshots/icon-explorations-02/d-detail.png)|[查看](../../web/ui-screenshots/icon-explorations-02/d-desktop.png)|[查看](../../web/ui-screenshots/icon-explorations-02/d-launcher.png)|
|E 色域|全幅纯色底、白色与彩色块面|[查看](../../web/ui-screenshots/icon-explorations-02/e-detail.png)|[查看](../../web/ui-screenshots/icon-explorations-02/e-desktop.png)|[查看](../../web/ui-screenshots/icon-explorations-02/e-launcher.png)|
|F 开放|无底板、独立双色图形|[查看](../../web/ui-screenshots/icon-explorations-02/f-detail.png)|[查看](../../web/ui-screenshots/icon-explorations-02/f-desktop.png)|[查看](../../web/ui-screenshots/icon-explorations-02/f-launcher.png)|
|G 叠纸|切角与折面、几何构图|[查看](../../web/ui-screenshots/icon-explorations-02/g-detail.png)|[查看](../../web/ui-screenshots/icon-explorations-02/g-desktop.png)|[查看](../../web/ui-screenshots/icon-explorations-02/g-launcher.png)|
|H 夜航|石墨底、彩色 Lucide 线条|[查看](../../web/ui-screenshots/icon-explorations-02/h-detail.png)|[查看](../../web/ui-screenshots/icon-explorations-02/h-desktop.png)|[查看](../../web/ui-screenshots/icon-explorations-02/h-launcher.png)|
|I 交织|无底板、几何拼接标记|[查看](../../web/ui-screenshots/icon-explorations-02/i-detail.png)|[查看](../../web/ui-screenshots/icon-explorations-02/i-desktop.png)|[查看](../../web/ui-screenshots/icon-explorations-02/i-launcher.png)|
|J 清透|透明浅底与边界、局部色片|[查看](../../web/ui-screenshots/icon-explorations-02/j-detail.png)|[查看](../../web/ui-screenshots/icon-explorations-02/j-desktop.png)|[查看](../../web/ui-screenshots/icon-explorations-02/j-launcher.png)|
|K 索引|彩色侧条、深色 Lucide 符号|[查看](../../web/ui-screenshots/icon-explorations-02/k-detail.png)|[查看](../../web/ui-screenshots/icon-explorations-02/k-desktop.png)|[查看](../../web/ui-screenshots/icon-explorations-02/k-launcher.png)|
|L 墨色|不对称底形、墨色主体与彩色切面|[查看](../../web/ui-screenshots/icon-explorations-02/l-detail.png)|[查看](../../web/ui-screenshots/icon-explorations-02/l-desktop.png)|[查看](../../web/ui-screenshots/icon-explorations-02/l-launcher.png)|

截图脚本 `web/.local/capture-icon-explorations.mjs` 只对开发服务器进行视觉预览，固定会话 API 并关闭测试 WebSocket，不接触真实节点。输出37张2倍像素截图。第一轮浏览器截图无 pageerror；视觉核对后修正 E 的浅色前景重叠和 K 的蒙版边缘不一致，并重新生成最终截图。`npm run check` 与 `npm run build` 通过；没有新增后端验收声明。B 的主体及底色渐变已移除，旧截图作为历史文件保留。
