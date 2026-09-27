# Compose 阶段输出（2026-09-12）

关联F09/F11/E02，整项仍进行中。

Compose执行阶段调用返回后，stdout/stderr分别持久化到操作日志，包括成功、失败和调用错误。每流保留前32 KiB并独立标记截断，最多16个阶段；操作日志写入前落实与读取一致的8 MiB预算。输出保存发生在后续Engine结果检查之前；保存失败保持之前Unknown记录，不凭CLI退出结果报告部署成功。配置归一化JSON输出没有作为执行日志公开。

容器query新增operation-output、taskId和零起点outputOffset。每次读取一个阶段，nextOutput=-1表示当前没有后续保存项。stdout/stderr用base64传输，避免任意控制字节的JSON膨胀；阶段最大响应保持96 KiB以内。Master与Daemon现有host.manage门禁生效，负页码在Master返回400；普通operation摘要不携带输出正文。

专用Compose任务详情增加输出浏览，解码后按纯文本显示，保留阶段页和截断标识；读取错误时隐藏缓存正文。刷新不重新执行命令。

验证：

- 独立CLI测试进程走实际runCLI，输出40 KiB、中文以及含NUL的stderr；成功/失败状态、双流截断、最大响应、拒权与分页、新管理器从磁盘读取，最终定向race2.120s通过。首次两次测试分别缺少Endpoint和/version协商而失败，补齐测试合同后通过。
- containers全包race6.004s通过（真实Engine条件入口未启用）；其后补充新管理器和双满流读取的定向结果如上。
- Master真实TLS host.manage及负页码拒绝，最终race1.917s通过。
- 本地Docker29.4.1、既有隔离镜像blora-isolated-e2e:20260909、真实HTTPS双Daemon fixture-2005453535：`npm run test:e2e:real -- tests/real/docker.spec.ts --workers=1 --reporter=line` 1/1（45.9s）通过。保存/显式部署/实际输出分页、任务详情显示、刷新恢复及删除部署/独立测试卷清理均通过；fixture已Ctrl+C停止。
- 前端41/41、生产构建、make check/build/windows通过；OpenAPI3.1解析103路径通过。最终说明文案重新构建结果见PROGRESS。

后续增量：命令启动前写入未完成输出记录，运行期间每250毫秒检查双流，仅正文或截断标记变化时原子落盘。输出仍各限32 KiB，达到上限后不会反复写相同内容。日志由执行协程独占修改，CLI写入有锁缓冲区；退出后写completed=true，再执行Engine资源检查。completed=false/缺省表示尚无退出记录，不能据此声称CLI仍活着。界面已显示该区别。

检查点保存失败会取消并等待受控CLI退出后返回，保留Unknown，不继续检查并报告成功。新增真实CLI进程测试：命令保持阻塞时由另一管理器读磁盘输出、取消后保留完成/错误/不明语义；用临时目录结构故障使后续原子写失败，确认CLI终止、旧检查点完整且未完成。定向race2.470s通过，容器全包race6.520s通过；随后追加目录故障测试的定向race1.952s通过。`make check build windows`及前端生产构建退出0。

Linux执行器SIGKILL追加证据：`go test -race ./internal/containers -run '^TestComposeOutputSurvivesExecutorSIGKILL$' -count=1` 最终1.381s退出0。测试从独立执行进程获得运行中检查点，再SIGKILL该执行进程；新管理器读取逐字节相同的未完成输出，operation呈现INTERRUPTED/Unknown，同任务Execute返回ErrUnknown、不重放命令。测试CLI按其独立会话与随机令牌清理。这是容器执行器级硬崩溃证据，未替代完整Daemon启动恢复。

真实Docker浏览器复验：fixture-4124677712（HTTPS双Daemon、本地Docker），同一docker.spec.ts增加completed与“命令已退出”断言，1/1（45.9s）通过；部署和独立卷在测试内清理，fixture已Ctrl+C停止。OpenAPI3.1仍为103路径。

250毫秒为采样周期，不保证硬崩溃前最后一个采样间隔内的字节均持久化；磁盘阻塞也可能延长周期。Linux遗留CLI目前依靠运行进程中的会话令牌取消，Daemon启动路径尚未持久识别并处理遗留Compose CLI，是下一项代码缺口；本测试的专属清理不能算成产品恢复实现。物理ENOSPC/掉电、Windows真实CLI、远程Engine与完整F09/E02组合不计已通过。
