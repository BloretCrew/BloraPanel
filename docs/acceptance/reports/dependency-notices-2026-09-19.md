# 发行依赖通知

关联E09及技术规划12章依赖许可材料要求。此前发行脚本没有统一保留上游许可证/通知。现新增`scripts/third-party-notices.py`，`make package`在二进制/前端/SDK构建后、归档前执行收集。

依据六个Go发行二进制的`go version -m`读取实际模块及版本，从本机模块缓存读取原文；另保留Go安装包LICENSE/PATENTS。前端读取`web/package-lock.json`中已安装且非dev的包，核对安装版本与锁文件一致，保留包内LICENSE/NOTICE/COPYING/Copyright文件。此范围包含被项目列为dependencies的构建工具，不声称全部都被嵌入浏览器。未安装的平台可选包单独列出，不作为已分发对象。

两个npm包没有附许可证原文：`@xterm/addon-serialize@0.14.0`按其package.json中的确切commit核对[上游许可证](https://github.com/xtermjs/xterm.js/blob/f447274f430fd22513f6adbf9862d19524471c04/LICENSE)，与同源xterm包原文一致；`@vue/devtools-api@6.6.4`保留[确切版本的上游原文](https://github.com/vuejs/vue-devtools/blob/v6.6.4/LICENSE)。Rolldown平台binding使用同版本父包附带的LICENSE和第三方通知，版本不一致拒绝生成。Go模块保留modernc/libc的第三方通知、SQLite附加声明及wazero NOTICE等，未仅抄元数据中的许可名称。

实际命令`GOMODCACHE=/tmp/blora-go-mod python3 scripts/third-party-notices.py`退出0，104条记录（17个Go模块、Go安装包、86个npm包），46个未安装可选包另列。生成原文490,143B，SHA-256 `ef77052c9af3598f1818fb89452baa122d423ff688e15fffcf18e3d10a81f453`。最初`go list -m all`因未缓存的非发行依赖触发受限网络失败，改为读取实际二进制模块；首次生成因Fedora拆分GOROOT许可证失败，现支持实际系统许可证路径和显式配置。

`scripts/package.py`将通知纳入所有六类归档：Master/Daemon原有docs树、独立web和SDK根目录通知。rc3实际构建和六包完整MANIFEST校验通过，六包中通知与生成原文逐字节一致，见[rc3报告](release-rc3-2026-09-19.md)；rc2不包含此改动。不为项目代码擅自添加授权条款，也不把原文收集等同于对任意后续改动/分发场景的法律结论。
