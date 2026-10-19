# 修复独立复核

日期：2026-10-15。复核对象：`lightpanel-refactor/lightpanel-main` 当前修复源码。复核角色：Arbiter。此记录复核实际实现和回归结果，不表示对全部运行环境作安全保证。

## 结论

本轮审计发现对应的主要修复已形成实际代码及拒绝/恢复行为测试。独立审阅未发现下载取消与提交之间的互锁死锁；暂停/删除会取消并等待原 worker，完成提交后删除任务记录不会删除应用。旧审计文件记录的是修复前结果，不能作为当前版本仍有相同问题的证据。

复核过程中另发现并实际复现了两项编辑回归：重命名后工作目录仍指向旧路径，以及显式传入旧路径时未对最终重命名目标检查安装占位。实现者已修复，两项独立回归测试现均通过。还反馈了自动识别命令应相对于所选工作目录，以及仅修改路径时应验证目录存在；最终代码已按这些建议补充。后两项为代码复核，Linux 可执行文件识别未作实机验证。

## 覆盖

| 范围 | 复核结果与证据 |
|---|---|
| 安装所有权和失败清理 | 统一安装计划、规范名称及路径重叠检查、进行中占位、唯一 staging；HTTP 错误、超限、短读、错误归档、配置失败均不得提交。已有文件保留与 null 配置释放锁由正常回归测试验证。 |
| 取消、暂停、恢复、提交 | 阅读 `install_jobs.go` 所有加锁分支。action 释放 job mutex 后才等待 done；恢复等待旧 worker 退出；最终提交与取消使用同一 job mutex 串行化。上下文取消后不提交、暂停后恢复、删除已完成记录不删除应用均有动态测试。未使用 race detector，不声称消除了所有竞争。 |
| 商店源和元数据 | 缓存绑定源 URL，确认摘要绑定源与完整条目；普通安装与参数确认共享安装管线。独立测试通过两个实际 handler 验证过期摘要被拒绝。来源、包 URL、初始化命令、工作目录、图标、端口及版本在安装转换中保留。 |
| 出站 HTTP | 校验协议、地址与用户信息，每次重定向重新验证；DialContext 解析并检查全部结果，然后拨号至同一已验证 IP。公共 transport 不继承环境代理。重定向和私网目标拒绝有测试；未向真实内部服务发送请求，也未进行外部 DNS 基础设施测试。 |
| 初始认证和改密 | 随机初始密码或显式初始密码；PBKDF2 保存并兼容旧 SHA256 登录迁移；会话绑定当前账号凭据摘要。独立测试确认旧默认密码升级后仍受强制改密限制。改密使所有旧会话无效、会话写入失败不发 cookie 均有测试。 |
| 页面与分析 | 分析结果使用文本节点；上传页面附带不含 allow-same-origin 的 CSP sandbox；读写删除统一路径边界。动态测试检查页面边界和响应策略。没有浏览器端 CSP/XSS 实际执行验证。 |
| 备份恢复 | 排除策略在路径规范化后检查会话、初始密码及 PID；备份输出网络传输前释放应用锁；恢复/备份拒绝活动安装占位。配置结构验证和排除路径回归通过。恢复仍是逐文件替换，不是整个数据目录事务；磁盘故障中断的全局回滚未实现或验证。 |
| 进程管理 | 进程记录包含 PID 与 CreateTime，停止前重新核验身份，降低旧 PID 被复用时误杀风险。现有 Windows 辅助进程测试验证手动停止不重启和连续崩溃的重试上限。旧裸 PID 文件无法被新版本接管，升级需先停止旧版受管进程。 |
| 编辑/手动添加 | 最终目录先规范化并检查名称、目录重叠及安装占位；重命名对原应用内部工作目录重映射；配置保存失败时回退目录。两项独立重命名回归通过。 |

## 独立新增测试

文件：`handlers/regression_review_test.go`。

- `TestReviewRenameRebasesInternalWorkDir`：重命名后工作目录和磁盘路径一致。修复前失败，修复后通过。
- `TestReviewRenameRejectsReservedFinalTarget`：最终目标被安装占位时保留原应用和目录。修复前失败，修复后通过。
- `TestReviewLegacyDefaultLoginIsForcedToChange`：旧默认凭据升级后依旧限制管理操作，仅允许设置及改密。
- `TestReviewChangedManifestCannotStartInstall`：普通安装、参数确认均拒绝已改变清单的旧摘要。

## 执行结果

在 Windows、隔离临时目录与项目 Go 工具链运行：

```text
go test ./handlers -run TestReview -count=1 -timeout 30s
ok lightpanel/handlers 1.115s

go test ./... -count=1 -timeout 60s
?  lightpanel [no test files]
?  lightpanel/config [no test files]
ok lightpanel/handlers 17.384s
?  lightpanel/models [no test files]

go vet ./...
退出码 0，无诊断
```

GOPROXY=off，GOTOOLCHAIN=local；GOCACHE、GOMODCACHE、TEMP/TMP 均指向项目外的 `lightpanel-refactor/tools`。测试不使用现有用户数据，不下载或运行第三方应用。首次全量检查指出旧备份测试使用了不符合新增 schema 的 fixture；该 fixture 修正为有效 Project 对象后全量通过，新增结构拒绝策略保留。

## 边界

- 本复核没有运行 Linux 实机服务、systemd 安装或真实 Linux 进程组停止；跨平台编译不等同这些运行验证。
- 未运行 Go race detector；互斥关系结论来自代码审阅和受控取消/恢复测试。
- 浏览器交互及 CSP 执行、依赖漏洞扫描和最终发行包校验由主任务另行记录，本记录不冒充完成这些工作。
- 应用命令、初始化命令和备份仍应来自管理员信任的来源。应用目录不是操作系统隔离沙盒，本轮未增加容器或独立系统用户隔离。
- 复核者没有修改生产代码；仅新增独立回归测试和本记录。测试通过表示这些回归断言成立，不是“所有代码无缺陷”。
