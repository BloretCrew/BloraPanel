# WebKit 真实监控历史验收（2026-09-22）

## 环境

- 宿主：Fedora Linux 43；本机 Playwright WebKit 的 Ubuntu fallback 二进制缺少 ABI 依赖，不能直接启动。
- 浏览器：官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 隔离容器中的 Playwright WebKit 1.63；与项目 `@playwright/test` 版本一致。
- 服务：本轮启动的私有 HTTPS Master 和双 Daemon fixture；容器以 host network 访问该 fixture，并只通过 host PID namespace 找到测试脚本按配置路径验证过的 fixture Daemon。

## 命令

```sh
./dist/blora-devfixture -listen 127.0.0.1:9443 -state-parent /tmp

docker run --rm --network host --pid host \
  -e BLORA_E2E_CREDENTIALS="$FIXTURE_DIR/browser-credentials.json" \
  -e BLORA_BROWSER=webkit \
  -e BLORA_E01_HISTORY_SOAK=1 \
  -e BLORA_E01_SAMPLE_INTERVAL_MS=1000 \
  -v /data/instances/blora-panel:/work \
  -v "$FIXTURE_DIR:$FIXTURE_DIR:ro" \
  -w /work/web mcr.microsoft.com/playwright:v1.63.0-noble \
  npx playwright test --config playwright.real.config.ts \
  tests/real/monitor-history-soak.spec.ts --workers=1 --trace=off --reporter=line
```

## 结果

`real node sampling retains exactly the newest 120 points over sustained polling` **1/1 通过**；Playwright 的 `real-test-results/.last-run.json` 为 `{"status":"passed","failedTests":[]}`。

场景以 1 秒间隔取得 121 个真实节点指标，确认历史恰保留最近 120 点、时间严格递增；随后仅暂停该 fixture 的 Daemon，确认 API 以 `stale` 标记完整缓存；fixture supervisor 重启 Master 后，缓存仍可读；恢复 Daemon 后节点重新上线。该结果补齐 E01/F10 的 WebKit 实际浏览器链路。

测试结束后容器以 `--rm` 删除，fixture 收到 Ctrl+C 后停止；复核没有残留 Master、Daemon、WebKit 进程、9443 监听或 fixture 目录。

## 边界

这是 Linux 容器中的 WebKit 验证，不是 Windows 浏览器、24 小时墙钟、远端网络或高 RTT/丢包性能证明；这些条件继续按验收矩阵保持进行中。
