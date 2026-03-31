# Transfer CloudSpace

一个全栈云文件管理系统，支持登录注册、文件上传下载、目录管理、分享链接和个人 notes。

## 技术栈

- 前端：React + Vite + TailwindCSS
- 后端：Go + Gin
- 数据库：PostgreSQL
- 部署：Nginx + systemd + Docker（数据库）

## 仓库结构

```text
transfer/
├── frontend/        # 前端项目（Vite）
├── backend/         # 后端项目（Go）
├── docs/            # 部署/复盘文档
└── openapi.yaml     # API 定义
```

## 核心功能

- 账号登录/注册（JWT 鉴权）
- 文件上传、列表、下载、删除、移动
- 文件夹树、新建文件夹、重命名
- 分享链接创建、公开信息页、公开下载
- 个人 notes（创建、查询）

## 本地开发

### 1. 启动数据库（示例）

```bash
docker run --name transfer-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=transfer \
  -p 5432:5432 -d postgres:16
```

### 2. 启动后端

```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/server
```

默认健康检查：

```bash
curl http://127.0.0.1:8080/healthz
```

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev
```

## 生产部署（小规格 ECS 推荐流程）

推荐“本地构建 + 上传服务器”，避免 ECS 在线构建占满内存。

### 后端发布

```bash
# 本地
./backend/scripts/build_linux_amd64.sh
scp ./backend/transfer-server emono:~/transfer-server

# 服务器
sudo mv ~/transfer-server /opt/transfer/backend/transfer-server
sudo chmod +x /opt/transfer/backend/transfer-server
sudo systemctl restart transfer-server
sudo systemctl status transfer-server --no-pager
```

### 前端发布

```bash
# 本地
cd frontend
npm run build
COPYFILE_DISABLE=1 tar -czf dist.tar.gz dist
scp dist.tar.gz emono:~/dist.tar.gz

# 服务器
sudo tar -xzf ~/dist.tar.gz -C /opt/transfer/frontend
sudo nginx -t && sudo systemctl reload nginx
```

## Nginx 关键配置

如果前端已上传到 `/opt/transfer/frontend/dist`，`root` 必须指向该目录，否则会出现“部署成功但页面还是旧版本”。

```nginx
server {
    listen 80 default_server;
    server_name _;
    client_max_body_size 200m;

    root /opt/transfer/frontend/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /healthz {
        proxy_pass http://127.0.0.1:8080/healthz;
    }
}
```

## 常见问题

- `npm ci` 报错缺少 lockfile：先执行 `npm install` 生成 `package-lock.json`。
- 前端部署后没变化：检查 Nginx `root` 是否指向 `dist`，并清浏览器缓存。
- 上传报 `413 Request Entity Too Large`：增大 `client_max_body_size`。
- ECS 构建崩溃：优先本地构建后上传。

## 相关文档

- `docs/deployment-log-2026-03.md`
- `docs/full-journey-2026-03.md`
- `docs/transfer-project-experience-cn.md`

