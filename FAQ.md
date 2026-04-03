# FAQ / 常见问题

## 编译问题

### 1. `bash: go: command not found`

**原因：** 系统未安装 Go。

**解决：**
```bash
wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

### 2. `dial tcp 142.250.204.49:443: i/o timeout`

**原因：** 国内网络无法访问 `proxy.golang.org`。

**解决：**
```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### 3. `missing go.sum entry for module`

**原因：** `go.sum` 文件缺失（源码包中未包含）。

**解决：**
```bash
go mod tidy
```

### 4. `cannot refer to unexported field last`

**原因：** 模型字段首字母小写，外部包无法访问。

**解决：** 已在 v2.1.6 修复，将 `last` 改为 `Last`。

### 5. `cannot use p.MemoryPercent() (float32) as float64`

**原因：** gopsutil v3 的 `MemoryPercent()` 返回 `float32`，与模型 `float64` 不匹配。

**解决：** 已在 v2.1.6 修复，添加 `float64()` 类型转换。

### 6. `"net/http" imported and not used`

**原因：** 某些文件中导入了未使用的包。

**解决：** 已在 v2.1.6 清理所有未使用的导入。

---

## 部署问题

### 7. 如何一键编译？

```bash
bash build.sh
```

### 8. 编译后如何运行？

```bash
./lightpanel
# 访问 http://127.0.0.1:31956
# 默认账号: admin / admin
```

### 9. 如何后台运行？

```bash
nohup ./lightpanel > /dev/null 2>&1 &
```

### 10. 如何设置开机自启？

创建 systemd 服务：
```bash
cat > /etc/systemd/system/lightpanel.service << EOF
[Unit]
Description=LightPanel
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/lightpanel
ExecStart=/opt/lightpanel/lightpanel
Restart=always

[Install]
WantedBy=multi-user.target
EOF
systemctl enable lightpanel
systemctl start lightpanel
```

---

## 使用问题

### 11. 如何修改端口？

编辑 `config/config.go`，修改 `Port = "31956"` 后重新编译。

### 12. 如何修改默认密码？

登录后在设置页面修改，或直接编辑 `data/config/user.json`。

### 13. 应用启动失败怎么办？

首页会弹出失败弹窗，显示日志内容和可能缺失的依赖。

### 14. 下载管理中的任务如何清理？

已完成/失败的任务会在 30 分钟后自动清理。

### 15. 如何更换 UI？

所有页面 HTML/CSS/JS 集中在 `handlers/templates.go`，修改后 `go build` 即可。

---

## 已知限制

| 限制 | 说明 |
|------|------|
| 仅支持 Linux | 依赖 `syscall.Kill`、`Setpgid`、`/proc` 文件系统 |
| 单用户 | 仅支持一个 Basic Auth 账号 |
| 无 HTTPS | 面板绑定 `127.0.0.1`，需 Nginx 反代实现 HTTPS |
| 无数据库 | 配置存 JSON 文件，不适合大规模部署 |
