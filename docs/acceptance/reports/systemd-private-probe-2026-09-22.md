# 私有 systemd user manager 探针（2026-09-22）

本轮只探测是否能在临时目录内形成可连接的独立 user manager，不触碰宿主服务、宿主 cgroup 或宿主 D-Bus。

## 环境与命令

- Linux x86_64；`/usr/lib/systemd/systemd` 与 `systemctl` 均为 systemd 258。
- 通过 `unshare --user --map-root-user` 启动隔离 user manager：uid 映射被当前沙箱拒绝；加入 `/proc` 挂载时同样返回 `Operation not permitted`。
- 直接在随机 `XDG_RUNTIME_DIR=/tmp/blora-systemd-user.*` 下启动 `/usr/lib/systemd/systemd --user`，再执行 `systemctl --user is-system-running`。

## 结果

隔离启动没有得到可运行的 manager。直接启动的进程不能建立可连接的本地 user bus，`systemctl --user` 返回 `Failed to connect to user scope bus via local transport: Operation not permitted`；宿主 `systemctl --user` 与 `systemctl` 也都因当前运行环境的总线权限返回同类错误。没有可执行的 service/timer 生命周期，因此没有把这次探针计为 E05 通过。

探针结束后删除所有临时运行目录，并复核没有 `systemd --user`、`dbus-broker`、临时 socket 或 `blora-systemd-user.*` 目录残留。未修改任何宿主 unit、服务、cgroup 或防火墙规则。

## 结论

E05 的真实 systemd service/timer 生命周期仍需一个可操作的独立 systemd 环境；当前 runner 没有安全且可连接的本地替代路径。Linux 适配器边界、参数校验和受控替身测试证据继续有效。
