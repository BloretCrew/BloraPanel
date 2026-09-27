# 真实浏览器全套回归

2026-09-19 当前基础夹具组合：排除另需Docker/性能夹具的4个文件及已单独通过的nodes文件，余下40项38通过/2失败（6.3m）。快捷入口断言误把夹具已有启动历史作为本次副作用，改为比较操作前任务ID后，定向1/1（3.6s）通过。任务分页101次mkdir暴露控制接收队列突发断链，已修复；新夹具通知文件2/2（8.2s），加强全部101任务成功断言后再次2/2（7.9s）通过，详见[协议背压报告](control-backpressure-2026-09-19.md)。这些是失败后的定向复验，不将原40项运行改写为40/40。

修复后完整节点复验：同一私有夹具执行 `npm run test:e2e:real -- tests/real/nodes.spec.ts`，2/2 通过（53.1s，退出 0）。实际暂停所属 Daemon 后失联提示、快捷入口状态、WAITING_NODE 和恢复后原运行身份均通过，维护/票据场景再次通过。

2026-09-19 根因与修复：节点维护测试通过 `[data-app="blora.nodes"]` 查找桌面入口，但桌面仅显示前六个默认应用，节点管理应从 `.launcher-button` 打开的应用菜单进入。已修正测试入口，并为页面动作设置 10 秒超时、记录主流程错误和保留清理预算，避免 finally 掩盖首次失败。私有 HTTPS 双 Daemon `fixture-1890745423` 执行 `npm run test:e2e:real -- --grep 'real node configuration'`，1/1 通过（4.7s，退出 0）。实际覆盖配置草稿刷新、维护期间原进程存活、禁止重启但允许停止、配额拒绝、换钥和登记票据显式下载、票据不写入工作现场。该修复属于测试入口修复，不是节点后端故障；不证明 Windows 或远程长期场景。

2026-09-19 独立复验纠正：新建私有双节点 `fixture-4276306771` 后执行 `npm run test:e2e:real -- --grep 'real node configuration'`，仍在 60 秒超时失败。最终错误位于 `nodes.spec.ts` finally 中的节点列表查询，掩盖了主流程停留位置。该结果不支持此前“仅由前序测试污染造成”的解释；根因未确定，节点维护/票据下载完整浏览器场景仍未通过。本轮夹具已请求停止，下一步先改善分阶段诊断和清理超时，定位主流程阻塞点，不提升验收状态。

使用最终 Linux 构建、HTTPS 双 Daemon fixture `fixture-2323261112` 执行：

```sh
BLORA_E2E_CREDENTIALS=/data/instances/blora-panel/.local/fixture-2323261112/browser-credentials.json \
npm run test:e2e:real -- --workers=1 --trace=off --reporter=line
```

当时 43 项中 37 项通过。6 项失败记录如下（节点场景的后续定位见上文）：

- `container-metrics.spec.ts`：fixture 节点没有可用容器运行时/镜像，创建容器实例返回 `CAPABILITY_UNAVAILABLE`；
- `docker.spec.ts`：Docker 查询返回 502，当前环境未提供可用 Docker Engine；
- `performance.spec.ts`、`upload-close.spec.ts`：这两项明确要求以 `--performance` 启动 fixture，本次基础 fixture 未带该参数；
- `nodes.spec.ts` 的维护场景超时，最初归因为前序污染；后续独立复验排除该解释，最终定位并修复应用入口定位器；
- `settings-editor.spec.ts` 同样受该污染导致初次状态为 RUNNING，独立新 fixture 重跑 1/1（29.8s）通过。

其余默认应用、文件、编辑器、终端、传输、任务、节点、工作区、扩展和恢复场景均通过；fixture 已停止。上述失败不改写为通过，也不把缺少 Docker、性能 fixture 或 Windows 真机的环境条件伪装成产品证据。
