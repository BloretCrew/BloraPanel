# Linux 原生通知绘制与点击补验（2026-09-29）

使用当前RC26产品构建、真实HTTPS Master/双Daemon、Playwright 1.63.0的headed Chromium，以及一次性容器内独立Xvfb/DBus/Dunst。产品UI、通知实现和此前选定的配色/材质均未修改。

## 真实链路

1. 用户入口启用系统通知，真实节点`file.mkdir`任务完成后，Dunst实际显示一条原生通知。
2. 从X11根窗口截图，确认弹窗位于浏览器外，显示“任务成功”、测试来源和`file.mkdir`；不是网页截图或站内提示。
3. XTest鼠标点击Dunst窗口，应用打开绑定该taskId的专用任务窗口，只有对应任务卡片。
4. Dunst历史中的标题与正文正确，弹窗关闭；网页刷新保留任务窗口，不重发旧通知。
5. 第二个真实任务产生新弹窗后，关闭整个浏览器。当前Linux组合中，Dunst保留已投递弹窗；右键真实鼠标操作可将其关闭。

没有替换Notification、注入其click事件或通过D-Bus直接调用通知动作。Dunst左键按[官方FAQ](https://dunst-project.org/documentation/faq/#clicking-on-a-notification-with-a-message-like-click-here-does-nothing)配置为`do_action, close_current`。截图使用scrot读取实际X11桌面，点击使用xdotool的XTest输入。

## 入口和证据

```sh
docker build --network host -t blora-native-notifications:e2e -f scripts/native-notifications.Dockerfile scripts
python3 scripts/native-notifications-e2e.py --output .local/evidence/native-notifications-new-run
```

入口包含独立Playwright配置、`web/tests/platform/native-notifications.spec.ts`、私有显示/总线脚本和Dockerfile。输出目录必须新建；凭据只读挂载，不输出密码。容器通过host网络连接本轮loopback fixture，没有宿主显示、DBus、设备或Docker socket挂载。只有测试、截图、容器退出及所属fixture清理全部成功才报告PASSED；fixture提前非零退出也拒绝。

私有证据位于`.local/evidence/rc26`：

| 日志 | 实际结果 |
| --- | --- |
| `native-image-build.log` | 镜像构建成功，53.384秒；不安装宿主软件 |
| `native-notifications.log` | 默认9445端口不被fixture支持，0.497秒退出1，未运行浏览器；改为允许的9444 |
| `native-notifications-retry.log` | 弹窗/点击/刷新通过，12.625秒退出0 |
| `native-notifications-lifecycle.log` | “浏览器退出自动移除Dunst通知”的预期失败，19.030秒退出1；实际仍有1条通知 |
| `native-notifications-final.log` | 绘制、标题、点击、刷新、退出后保留及用户关闭通过，9.695秒退出0 |
| `native-notifications-cleanup.log` | 最终入口和X11重绘后截图复核通过，9.666秒退出0，清理成功 |

最终原生桌面PNG位于`native-notifications-cleanup/results/`，包含浏览器运行时及退出后两个场景。类型检查、Python/Shell语法和差异检查通过。全部所属容器与fixture进程已结束；日志与截图留在私有证据目录，不公开凭据。

## 边界

2026-09-20的证据只有真实Notification对象与脚本派发点击；本轮补齐 **Linux X11/Dunst实际绘制及系统鼠标点击**。

浏览器退出后自动消失不是现有F09合同，也未被本轮实现。测试按观察到的Dunst行为记录保留通知并验证用户关闭，不把原失败改写成自动消失通过。没有承诺浏览器退出后继续投递、点击后重启应用或恢复窗口；这些需要持久通知/后台机制，不属于当前已验证能力。

Windows、macOS、Wayland、GNOME/KDE通知中心和声音/免打扰策略仍未验证。虚拟显示中的真实通知进程和XTest不等于实体桌面人工验收。F09整项、严格E08和原Vim偶发问题不因本次子项通过而自动关闭。
