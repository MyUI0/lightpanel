# LightPanel

一个极简的 Linux 服务器自部署二进制项目管理面板。

> ⚠️ 本项目为**初始版本**，功能尚不完善，仍在持续迭代中。如有不足请见谅，欢迎反馈和贡献。
>
> 本项目由 AI 辅助编写，代码可能存在未发现的边界情况，请自行测试后部署。

## 特性

- **零依赖部署** — 单二进制文件，无需 Docker / systemd
- **应用商店** — 支持多源 JSON 拉取，一键部署
- **下载管理** — 异步下载、暂停/继续、断点续传、实时进度
- **进程守护** — 启停/重启/删除、开机自启、崩溃自动重启（最多 3 次/5 分钟）
- **日志管理** — 自动轮转、按错误/警告/信息/崩溃过滤
- **系统监控** — CPU/内存/磁盘/实时进程列表
- **启动失败检测** — 自动分析日志，提示缺失依赖
- **安全** — Basic Auth + POST 防护 + 路径遍历过滤 + 原子配置写入
- **模块化** — 14 个文件职责清晰，UI 集中在 `templates.go`，易于二改

## 适用场景

适用于管理单二进制或脚本类自部署项目，如：
- [AList](https://github.com/alist-org/alist) · [Memos](https://github.com/usememos/memos) · [哪吒探针](https://github.com/nezhahq/agent) · [frp](https://github.com/fatedier/frp) · [sing-box](https://github.com/SagerNet/sing-box) 等

## 快速部署

```bash
# 1. 下载源码
git clone https://github.com/你的用户名/LightPanel.git
cd LightPanel

# 2. 编译
go mod download
go build -o lightpanel .

# 3. 运行
chmod +x lightpanel
./lightpanel
```

默认访问 `http://127.0.0.1:31956`，账号密码 `admin / admin`。

## 项目结构

```
main.go              # 入口
config/config.go     # 配置常量
models/models.go     # 数据模型
handlers/
  auth.go            # 认证 + POST 中间件
  routes.go          # 路由注册
  utils.go           # JSON 读写 + 初始化
  templates.go       # 所有页面 HTML/CSS/JS（改 UI 只需改此文件）
  page_dashboard.go  # 仪表盘 + 系统监控
  page_store.go      # 应用商店 + 源管理
  page_settings.go   # 设置 + 日志 + 进程管理
  page_edit.go       # 编辑应用
  app.go             # 应用核心逻辑 + watchdog
  app_control.go     # 应用控制
  app_create.go      # 应用创建
  downloads.go       # 下载管理
```

## 修改 UI

所有页面集中在 `handlers/templates.go`，改完 `go build` 即可，无需构建工具。

## 免责声明

1. 本项目为个人学习/工具用途，作者不对使用本项目造成的任何数据丢失、系统故障或安全事件负责。
2. 本项目由 AI 辅助编写，代码可能存在未发现的边界情况，请自行测试后部署。
3. 请勿在生产环境中直接使用，建议先在测试环境验证。
4. 使用本项目即表示您同意自行承担所有风险。

## 许可证

MIT License — 可自由拉取、修改、二次发布，仅需保留原作者信息。
