# Arbiter 独立复核记录

日期：2026-10-15。对象：当前 `lightpanel-refactor/lightpanel-main`。未改生产代码；新增证据测试 `handlers/audit_arbiter_test.go` 受 `securityaudit` build tag 隔离。此记录用于复核主审计与 Aegis 的结论，不代表全部功能已通过安全认证。

## 验证方式

独立阅读了路由/鉴权、创建、下载任务、商店缓存、分析模板、恢复与路径检查、页面读写和自动识别实现。独立重跑已有 13 个 Audit 测试全部通过，再增加 3 个不同场景复现全部通过。这里的 PASS 表示缺陷复现成功，不表示修复。测试均使用临时数据目录及内存 HTTP transport，没有访问外网或运行第三方程序。

执行命令：`go test -tags securityaudit ./handlers -run Audit -v -count=1 -timeout 60s`。工具链为项目外的 Go 1.27.2，Windows；GOPROXY=off、GOTOOLCHAIN=local，TEMP/TMP 指向项目工具目录下的 audit-temp。未发现 gcc/clang，未运行 race detector；Linux 运行时及实际浏览器 XSS 执行不在本复核的动态证据范围内。

## 复核结论与严重度边界

| 发现 | 判定 | 建议优先级与边界 |
|---|---|---|
| 远程脚本依赖 token 经分析接口进入 `innerHTML` | 确认代码链及 token 动态结果；`<svg/onload=alert(1)>` 可完整到达 HTML sink | P1。管理员仅分析脚本即引入同源脚本，无需安装或执行脚本。JSON Unicode 转义无效，因为前端解码后重新解析 HTML。未在本复核执行浏览器事件。 |
| `demo app` 创建失败删除已存在 `demo-app` 目录 | 动态确认 | P1。正常管理员操作可直接损失已有数据，与恶意 shell 命令无关。清理必须局限于当前任务独占目录。 |
| 参数安装确认路径绕过普通入口的名称/URL/重复校验 | 动态及静态确认 | P1，有前提。须直接进入参数确认或两步间源/缓存变化；普通商店安装先有名称检查，不能称普通按钮无条件触发。仍是目录边界缺陷，非未认证 RCE。 |
| 公网 URL 302 到回环未重新检查 | 内存 transport 及标准 http.Client 重定向动态确认 | P2。管理员认证前提；DNS 未解析属于静态证据。脚本分析不返回任意完整正文，勿声称可直接读取所有内网数据。 |
| 商店缓存跨源共用、索引未绑定安装计划 | 动态及静态确认 | P2。显示与确认内容可能错配；恶意来源影响取决于管理员所添加来源。 |
| 暂停后仍创建、删除任务后后台仍部署 | 动态确认两个场景 | P2。新增测试通过真实 action handler 删除 UI 任务，worker 仍写包并创建应用。快速恢复多 worker 的具体竞争窗口仅静态推断。 |
| 超限截断仍 completed，商店未检查状态码/大小 | 截断动态确认，其余静态确认 | P2。大小检查在写入后执行，所以实际末块也可能使文件略超上限；不得声称所有分支都执行统一大小限制。 |
| 自动部署/恢复丢失 SetupCmd、来源、显式 WorkDir、下载标志 | 独立数据流确认 | P2。`checkAppUpdates` 按 source manifest URL 匹配，而新建 Project 未保存该 URL。任务恢复未携带 AutoExtract/MakeExec。 |
| 自动检测以系统最近修改文件推断新应用 | 独立代码确认 | P2 功能正确性风险。未模拟 Linux 全系统竞争，也未认定跨权限利用。 |
| JSON null 配置导致创建 worker 吞 panic 后永久持锁 | 新增测试动态确认 | P2（较高修复优先级）。`app_create.go:260-272` 持锁后 nil map 赋值 panic，`recover` 吞掉且没 Unlock；任务保持 creating。测试在 worker unwind 后主动解锁以恢复测试全局状态。损坏/缺失配置同类路径，但新增动态证据具体是 `null`。 |
| 缓存与 Last 字段同步不完整 | 独立代码确认 | P2。不能写作已由 race detector 证实，更不能推断已存在内存破坏。 |
| 改密后其它有效会话仍保留 | 动态确认 | P2。已有会话最长至 24h 过期，影响泄露会话撤销；不会让未知随机 token 认证。 |
| 恢复 `config/./sessions/token.json` 绕过 session 排除 | 动态确认，导入 token 可通过认证 | P1/P2 取决于备份信任模型。建议 P1 修复明确的保留路径策略绕过，但报告应写“恶意备份植入会话”。备份本可包含 Cmd/hooks，必须受信，因此不应宣称新增未认证 RCE。 |
| 备份响应慢写持有全局应用锁 | 动态确认 | P2。持锁持续到 io.Copy 结束；服务缺少 WriteTimeout。认证下载者或慢链路可阻止启动/停止等操作，非 unauth DoS。 |
| 上传 HTML 原样同源执行 | 动态确认原样响应及无 CSP sandbox；执行行为由浏览器语义确定 | P2 架构/信任边界风险。HTML 功能可能有意允许脚本，必须明确其等同受信管理代码，或隔离源/iframe sandbox。不要与脚本分析页面无意 DOM XSS 混为同等性质。 |
| 页面路径前缀判断接受兄弟目录 | Windows 真实路由 GET 读取及 POST 删除动态确认 | P2 限 Windows。编码反斜杠可跨到 pages-other，删除不限 HTML。Linux 发布目标上当前 Go1.21 ServeMux 兼容行为会 301 清理编码斜杠，尚未证实 Linux HTTP 逃逸；不能直接报 Linux 任意文件读取。 |
| 无 cookie 访问 logout panic | 动态确认 | P3。net/http 恢复请求级 panic；不会因此直接退出服务器进程。 |

默认 admin/admin、初次登录仅提示而不强制更改、默认全接口监听应在部署前优先处理；其实际暴露面取决于防火墙/反向代理/绑定配置。单轮 SHA256 密码存储、没有 Secure cookie 和默认 HTTP 属于独立认证/传输加固事项；不能将 HttpOnly cookie 误称为可被 JavaScript 直接读取。

## 新增证据

1. `TestAuditArbiterNullConfigRetainsLifecycleMutex`：真实创建 worker 在 JSON null 配置下持锁 panic；响应 body 的 deferred Close 证明 worker 已展开退出，锁仍持有，任务仍 creating。
2. `TestAuditArbiterDeletedDownloadStillCreatesApp`：响应到达前调用实际 `/dl/action/<id>/delete` handler，确认 task 从 sync.Map 消失，但创建最终 completed、下载文件仍存在。
3. `TestAuditArbiterWindowsPageDeleteEscapesSibling`：通过 SetupRoutes、有效 session/CSRF POST `/page/del/..%5cpages-other%5cmarker.txt`，兄弟目录内临时文件确实删除。

## 不应作为高危结论的内容

- 管理员设置 Cmd/SetupCmd 本来就是管理功能；没有额外跨信任边界证据时，不单独报命令执行漏洞。
- 受管进程和面板同系统用户运行时，目录不是操作系统沙盒；符号链接或本地写文件需结合真正低权限攻击者边界论证。
- 归档处理已有边界、符号链接、大小与条目数防护；本复核未证实普通 tar/zip 解压越界。
- 不把静态竞争怀疑写成已复现崩溃，不把依赖版本旧等同存在已知 CVE。
- 本次为审计，尚未修复；修复后应把正向“缺陷存在”证据改成拒绝/恢复行为断言，再运行回归与 Linux 集成验证。
