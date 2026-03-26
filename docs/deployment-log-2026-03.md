# Transfer 项目部署与故障复盘（2026-03）

## 1. 本次完成结果
- ECS (Ubuntu 22.04) 可通过 SSH 稳定登录。
- 后端服务已在服务器启动并通过健康检查。
- 公网可访问：`http://8.163.53.203/healthz` 返回 `{"status":"ok"}`。
- 文件上传、分享链接、公开下载链路本地联调通过。
- 新增了 `notes` 接口（文本消息流）。

## 2. 后端功能落地内容
- 登录：`POST /api/auth/login`
- 文件：
  - `POST /api/files/upload`
  - `GET /api/files`
  - `GET /api/files/{fileId}/download`
  - `DELETE /api/files/{fileId}`
- 分享：
  - `POST /api/shares`
  - `GET /api/s/{token}`
  - `GET /api/s/{token}/download`
- 文本消息：
  - `POST /api/notes`
  - `GET /api/notes`

## 3. 数据库与存储
- 数据库：PostgreSQL（Docker 容器 `transfer-db`）
- 端口映射：`15432 -> 5432`
- 文件存储：`/opt/transfer/backend/uploads`（磁盘）

## 4. 关键问题与解决过程

### 问题 A：Docker 拉镜像/构建超时
- 现象：拉 `docker.io` 相关资源超时。
- 发现方式：`docker compose build` 报 `DeadlineExceeded` / `oauth token timeout`。
- 处理：配置 Docker 镜像源（镜像加速器）并重试。

### 问题 B：服务器内存小导致构建卡死，SSH 不稳定
- 现象：SSH 卡在 banner、远程连接不稳定。
- 发现方式：连接阶段超时，重启后短暂恢复。
- 处理：
  - 增加 2G swap。
  - 改用“本地编译二进制 + 上传服务器”方式，避免在 2G 机器上长时间 `docker build`。

### 问题 C：后端启动失败（连错数据库）
- 现象：`failed to initialize server: connect postgres ... localhost:5432 refused`。
- 发现方式：查看 `server.log`。
- 处理：在 `.env` 明确设置
  - `DATABASE_URL=postgres://postgres:<password>@127.0.0.1:15432/transfer?sslmode=disable`
  - 并确保 `transfer-db` 容器运行。

### 问题 D：Nginx 启动失败（80 端口冲突）
- 现象：`nginx bind() to 0.0.0.0:80 failed`。
- 发现方式：`systemctl status nginx` + `ss -lntp | grep :80`。
- 根因：旧容器 `file-transfer-app` 占用了 80 端口。
- 处理：停止并删除该容器后，Nginx 成功启动。

### 问题 E：Nginx 404 / default_server 冲突
- 现象：访问 `127.0.0.1/healthz` 返回 Nginx 404，或 `duplicate default server`。
- 发现方式：`nginx -t`。
- 处理：删除默认站点链接并保留单一反向代理配置。

## 5. 当前建议的安全组策略

### 入方向保留
- `SSH(22)`：来源建议改为你的当前公网 IP（不要 `0.0.0.0/0`）
- `HTTP(80)`：`0.0.0.0/0`
- `HTTPS(443)`：`0.0.0.0/0`（后续上证书时使用）

### 入方向删除
- `RDP(3389)`（Linux 不需要）
- `ICMP 全放行`（可选，不必长期开放）
- 所有“全端口放行”临时规则

> 不是留 `88`，是留 `80`（HTTP）。

## 6. systemd 托管方案（推荐）

### 6.1 安装服务文件
```bash
cp /opt/transfer/backend/deploy/transfer-server.service /etc/systemd/system/transfer-server.service
systemctl daemon-reload
systemctl enable --now transfer-server
```

### 6.2 查看状态
```bash
systemctl status transfer-server --no-pager
journalctl -u transfer-server -n 100 --no-pager
```

### 6.3 重启与停止
```bash
systemctl restart transfer-server
systemctl stop transfer-server
```

## 7. 密钥与密码安全事件
- 在沟通中曾暴露过 `POSTGRES_PASSWORD` 与 `DEMO_PASSWORD`。
- 已建议立即轮换所有暴露过的口令。
- `.pem` 私钥应仅存放于 `~/.ssh/`，不得提交到仓库。

## 8. 后续计划
- 上线 HTTPS（80 -> 443）。
- 前端接入并通过 Nginx 统一路由（`/` 前端，`/api` 后端）。
- 增加基础测试脚本（登录、上传、分享、notes）。
- 逐步引入 CI/CD（GitHub Actions）自动构建与发布。
