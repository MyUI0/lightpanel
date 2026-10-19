# Aegis 下载、商店与脚本分析审计

审计日期：2026-10-15。对象为当前重构源码。未改动生产代码。`handlers/audit_agent_aegis_test.go` 仅在 `securityaudit` build tag 下编译，全部 7 个正向复现通过；此处 PASS 表示复现了缺陷，不表示安全回归通过。复现仅使用临时目录和内存 HTTP transport，没有向公网或内网发包，没有执行下载代码。

优先级：P1 应先修复，P2 应修复，P3 改进。所有安装、下载动作在实际路由中受管理员登录和 CSRF 保护；脚本分析接口受管理员登录保护。本报告不把管理员自愿设置启动命令或信任安装脚本本身认定为命令注入。

## A01 / P1 — 分析远程脚本会触发 DOM XSS

- 位置：`handlers/web/templates/analyze.html:48`；输入来源 `handlers/script_analyze.go:68-73`。
- 触发：管理员分析攻击者控制的脚本，脚本包含 `apt install <svg/onload=alert(1)>`。分析接口把尖括号和事件属性完整保留为依赖项，前端解码 JSON 后直接拼入 `innerHTML`。HTTP JSON 的 Unicode 转义不能防止随后 DOM HTML 解析。
- 影响：攻击者代码以面板同源运行，可读取可读 CSRF cookie、代管理员请求控制接口；HttpOnly session cookie 不阻止同源请求使用会话。无需安装或执行脚本。
- 验证：`TestAuditAegisAnalyzerReturnsExecutableMarkup` 验证 payload 完整进入依赖结果；浏览器执行由主审计/独立验证者补充，未在此子审计运行浏览器。
- 修复：使用 `createElement` 和 `textContent` 创建依赖标签，或严格转义插入文本；检查其它动态 HTML 接口。

## A02 / P1 — 创建名称归一化碰撞，可删除已有应用数据

- 位置：`handlers/app_create.go:105-113,128-143`；其它失败清理在 `165,173,188,204,222`。
- 触发：已有应用 `demo-app`，创建名称 `demo app`。查重使用原名称，实际目录把空格转为连字符；两者共享同一目录。此后无效 URL、网络错误、非 200 或写入错误直接 `RemoveAll(sandbox)`。
- 影响：旧应用整个目录被删除，旧配置仍保留。是正常管理员操作可触发的数据损失，无需攻击者或管理员误授权删除。并发创建同名也没有保留名称/目录，能造成同类互相覆盖/清理。
- 验证：`TestAuditAegisCreateNameCollisionDeletesExistingDirectory` 已在临时目录重现，无效 URL 即删除旧目录。
- 修复：统一规范名称后再查重；对最终规范路径及进行中任务原子占位。下载到新建且唯一的 staging 目录，只清理当前任务真正拥有的目录，成功后原子提交。

## A03 / P1（有前提）— 参数确认接口漏校验导致目录穿越

- 位置：`handlers/downloads.go:695-706,743-763`，比较普通安装入口 `293-343`。
- 触发：已登录管理员直接进入某源条目的 `/install/params/<index>` 并确认，条目名称如 `../outside`；或从正常条目进入参数页面后源内容/全局缓存改变，确认时同一索引变为带穿越的名称。确认接口只检查名称非空，不复用普通安装入口的名称、URL、私网、重复安装校验。
- 影响：下载文件写入 AppsDir 外，已有同名文件可被截断覆盖。还可直接使用私网 URL，重复安装可以覆盖既有应用配置/文件。范围受面板进程的文件权限限制。
- 重要边界：普通 `/install/` 路径会先拒绝含 `..` 的名称，因此不能声称“在普通商店卡片点击一次即可触发穿越”；需要直接参数确认路径或两步之间的数据变化。管理员本来就能执行命令，本项是下载/安装目录边界缺失，不是新增未认证 RCE。
- 验证：`TestAuditAegisConfirmAllowsTraversalAndPrivateURL` 直接调用实际确认 handler，缓存置为模拟恶意源；确认在临时 DataDir 的 AppsDir 外创建文件，且请求 loopback URL。transport 在部署阶段前置 paused，未执行任何代码。
- 修复：共用一个不可变安装计划/校验函数，确认页面绑定源 ID、条目 ID、manifest 摘要及已校验字段；最终路径使用规范化边界检查，拒绝已存在目录/记录，不能信任前一步校验。

## A04 / P2 — SSRF 拦截只看 URL 文本，DNS 和重定向可绕过

- 位置：`handlers/security.go:24-48`；`handlers/script_analyze.go:34-39`；`handlers/app_create.go:141-160`；`handlers/downloads.go:341-392`；`handlers/page_store.go:42-56,118-119`。
- 触发：传入外观正常的域名但其 A/AAAA 指向私网；或通过检查的公网地址返回 302 到回环/私网。`isPrivateIP` 不解析 DNS，所有 `http.Client` 使用默认重定向，没有检查实际拨号 IP。`::` 等 unspecified 地址也未全面拒绝。
- 影响：管理员分析/下载远程内容时，远端可诱导面板向内网 GET；下载路径还可落盘内部响应。脚本分析只返回正则提取字段，不是完整响应读取接口。鉴权仍然需要，不应标为未认证 SSRF。
- 验证：`TestAuditAegisAnalyzerFollowsRedirectToLoopback` 使用标准 http.Client 重定向逻辑 + 内存 transport，证实第二次请求到 `127.0.0.1` 未重新验证。DNS 部分为代码推断。
- 修复：统一受控 transport，解析并验证每次实际连接的全部目标 IP，防止 DNS rebinding；每次重定向重新校验，明确处理代理和允许协议。分析增加总超时和请求 context，目前 `http.Get` 无超时，慢响应可无限占用请求。

## A05 / P2 — 商店缓存未按源区分，选源/安装的内容可能不同

- 位置：`handlers/downloads.go:223-230,254-258`；安装按缓存索引取条目 `281-292,689-695`。
- 触发：先打开源 A 填充缓存，再切换源 B。5 分钟内直接返回 A 的 apps，而不比较 srcIdx/URL。页面显示选中 B，内容和确认安装读取的仍是全局缓存。另一个页面/管理员刷新源也可能改变已打开页面索引的含义。
- 影响：部署错误来源/错误应用；与“确认安装”仅传索引叠加，可能安装与此前浏览的条目不同的代码。
- 验证：`TestAuditAegisStoreCacheIgnoresSource` 重现请求 B 返回 A 缓存。
- 修复：按稳定源标识分别缓存，安装按稳定应用 ID 和 manifest 版本绑定；对比确认展示内容与实际安装计划。

## A06 / P2 — 暂停、恢复、删除并非可靠的任务控制

- 位置：`handlers/app_create.go:195-233`；`handlers/downloads.go:929-944,1039-1068`。
- 触发：创建下载过程中点击暂停。createApp 循环不读取 task 状态，照常下载并把 paused 覆盖为 completed。商店下载删除操作仅从 sync.Map 删除，不改变运行中的 task 状态、不取消 HTTP、不等待 worker，后台仍可部署。商店快速 pause→resume 在旧 worker 退出前启动新 worker，可出现同文件双写。
- 影响：用户以为停止的下载继续占资源或部署；恢复可能破坏文件。createApp 名称含空格时，恢复/删除还使用 task.Name 原名，与实际连字符目录不一致。
- 验证：`TestAuditAegisCreateIgnoresPause` 重现暂停后仍创建应用；删除/恢复竞态由状态机代码推断，未运行时竞争验证。
- 修复：每个任务单一 worker、context cancellation、受锁状态迁移，暂停/删除等待 worker 完成；恢复同 worker 或确保唯一所有权；task 存最终路径，不重新推导。

## A07 / P2 — 下载和解压失败可被视为安装成功

- 位置：`handlers/app_create.go:227-241,263-275`；`handlers/downloads.go:392-440,489,753-801,850,993-1005,1067-1126`。
- 触发：创建下载超过 MaxDownBytes 时循环 `break`，继续标 completed，保存截断文件。商店安装和恢复完全没有下载大小上限，也不拒绝 404/500 响应。要求自动解压时 extractArchive 返回 false，调用者仍继续保存配置并报 completed。WriteJSON 返回值也普遍忽略。
- 影响：错误页、截断包、残缺解压结果被作为完成安装保留；恶意或错误服务器可耗尽磁盘（商店路径）。大文件上限约束与 UI 状态不真实。
- 验证：`TestAuditAegisCreateTreatsLimitAsSuccess` 将限额缩小到 16 bytes，24-byte 响应截断到 16 后仍 completed；其它分支静态确认。
- 修复：统一下载管线，校验状态码、实际字节数和限额，区分下载/校验/解压/提交，任何失败回滚独占 staging；显式处理保存错误。解压已有条数/尺寸和路径校验，但不能只返回 bool 且忽略失败。

## A08 / P2 — 自动部署与恢复丢失应用安装元数据

- 位置：`handlers/downloads.go:365-377,456-485,727-738,817-846,1071-1077,1095-1122`；更新检查 `625-629`。
- 触发：商店条目依赖 SetupCmd、WorkDir、自动解压/可执行标志或更新来源。任务创建未存 AutoExtract/MakeExec，正常自动部署未把 SetupCmd/SourceURL/显式 WorkDir 写入 Project；参数确认分支还没将 icon/port 存入 task。恢复检查空的 AutoExtract/MakeExec，行为变为不解压/不 chmod。
- 影响：需要初始化的应用无法启动；暂停前后不同部署结果；`checkAppUpdates` 要求 Project.SourceURL 匹配商店源 URL，但正常商店部署没存该字段，更新发现长期为空。手动创建的 SourceURL 是包 URL，也不等于源 URL。
- 验证：代码数据流确认，未实际运行第三方应用。
- 修复：统一 InstallPlan 和 Project 转换函数；分离 source_manifest_url 与 package_url，保持完整元数据和下载选项；测试无暂停和恢复后部署结果一致。

## A09 / P2 — 自动识别扫描全系统，可能选择无关程序

- 位置：`handlers/detect.go:82-158,161-188`；调用 `handlers/downloads.go:453-474`。
- 触发：安装期间其它操作在 `/usr/bin`、`/root`、`/home` 等创建/更新可执行文件。combinedDetect 优先选整个系统“最近更新”的候选，直接返回 high confidence，不要求候选属于新安装目录，也不要求脚本确认。常规解压包只放 sandbox，但该函数反而不优先找 sandbox 内的程序。
- 影响：WorkDir、默认启动命令可指向其它应用，启动错误程序；扫描成本随 home/opt 文件数量增长。并非已验证的跨权限漏洞，因为默认面板和受管应用常同用户运行。
- 验证：代码推断；未扫描本机敏感目录或创建可执行程序。
- 修复：优先限定已下载文件清单/安装目录；外部安装目录仅在明确安装结果或管理员选择后采用，不自动将时间相关性当归属证明。

## A10 / P2 — 未处理损坏配置可能吞 panic 并永久持有全局锁

- 位置：`handlers/app_create.go:260-272`；`handlers/downloads.go:312-332,1152-1161`。
- 触发：apps.json 丢失、无效或 JSON `null`，`LoadJSON` 失败/返回 nil map 后仍赋值。createApp 获得 appOpMu 后向 nil map 赋值 panic，外层 `recover()` 吞掉，Unlock 未执行。普通无下载 URL 商店入口也存在同样模式。
- 影响：任务永远停在 creating，之后应用操作等待全局锁，重启前无法正常管理。
- 验证：静态确认；未执行会锁死测试进程的复现。
- 修复：所有写操作先验证配置加载，初始化 nil map；持锁立即 defer Unlock；panic recovery 必须记录并设置失败状态，不能代替错误处理。

## A11 / P2 — 缓存和任务时间字段存在未同步读写

- 位置：`handlers/downloads.go:123-128,140-143,395,414,428`；`models/models.go:126-134`（UpdateProgress 更新 Last）。
- 触发：启动后后台 initStoreCache 未持 storeCacheLock，和打开商店并发读写；清理 ticker 无锁读 task.Last，下载进度在互斥锁内写，错误路径又直接写 Last。downloads 模板接收 live task 指针后也直接读取正在写入的字段。
- 影响：Go data race，清理与显示状态不可靠。未声称已复现崩溃或内存破坏。
- 验证：代码审查，未运行 race detector；独立验证者需有支持 CGO/race 的工具链。
- 修复：缓存所有入口遵守同一锁约定；时间字段只能通过同步方法访问，模板使用一次性 snapshot。

## 覆盖与边界

逐行覆盖 downloads.go、app_create.go、page_store.go、script_analyze.go、extract.go、detect.go；补充审查 security.go、archive_paths.go、models.go、routes.go、analyze/downloads/store/store_params 模板、安装脚本、hooks/notify、相关启动函数。认证、恢复、页面上传、进程生命周期由其它审计任务负责。

未将以下内容独立报为可利用漏洞：管理员自定义 Cmd/SetupCmd 的执行；管理员主动添加受信自托管 Bark URL；仅检查 ELF 魔数不能证明二进制安全（函数命名/供应链信任需说明）；install.sh 未验签、商店允许 HTTP、第三方清单无完整性 pin 属于供应链加固需求，尚未证明具体供应链受损。未对未查阅漏洞数据库的依赖编造 CVE。

提取器已对 tar/zip 路径穿越、现有符号链接、硬链接、条数和解压后声明尺寸设置检查；仍建议独占 staging 和原子提交，避免部分解压留在正式应用目录以及检查/使用路径之间竞态。此次未复现归档越界，不将猜测列为独立高危。
