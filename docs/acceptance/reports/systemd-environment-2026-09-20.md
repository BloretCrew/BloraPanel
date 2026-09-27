# systemd 运行环境复核（2026-09-20）

本轮检查了可安全使用的 Linux 环境：宿主提供 `/usr/bin/systemctl` 和 `/usr/lib/systemd/systemd`，但 `systemctl --user is-system-running` 无法连接本地 user bus，返回 `Operation not permitted`；直接启动 user manager 也不能形成可连接的管理总线。

为避免触碰宿主服务，使用一次性 Debian `bookworm-slim` 容器安装 systemd/dbus，并以不挂载宿主 cgroup、仅临时 `SYS_ADMIN` 能力的方式启动。容器立即以退出码 255 结束，未形成可操作的 systemd manager。带宿主 cgroup 挂载的特权容器方案被安全审查拒绝，未执行。

因此 systemd service/timer 的真实生命周期仍标为环境阻塞；已有 Linux 适配器边界、能力探测、systemctl 参数校验及受控替身测试证据继续有效。本次未修改宿主服务或 cgroup。
