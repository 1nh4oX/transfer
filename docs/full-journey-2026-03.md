# Transfer 项目从 0 到上线复盘（2026-03）

## 项目目标
构建一个跨平台文件传输工具，支持：
- 账号登录/注册
- 文件上传、下载、删除
- 分享链接生成与公开下载
- 目录/文件夹管理
- 右侧速记对话框（notes）
- 前端部署到 ECS，用户通过公网 URL 访问

## 一、基础设施阶段（阿里云 ECS）
### 做了什么
- 领取并创建了自己的 ECS（Ubuntu）。
- 通过 `.pem` 密钥从本地终端 `ssh` 登录服务器。
- 开通安全组入方向规则（至少 `22/80`，后续可扩展 `443`）。

### 典型问题
- `ssh timeout / banner exchange timeout / connection refused`
- 登录方式混淆：控制台远程连接 vs 本地 SSH

### 解决方式
- 确认安全组放行 `0.0.0.0/0` 的 `22`（开发阶段）。
- 确认 `sshd` 运行：`systemctl status ssh`、`ss -lntp | grep :22`。
- 使用 `chmod 400 xxx.pem`，本地 `ssh -i` 连接。

## 二、后端初版（Go + Gin + PostgreSQL）
### 做了什么
- 选择 Go 技术栈。
- 按 OpenAPI 先行设计接口，再实现后端。
- 建立了分层结构：`handler / service / repo / model`。
- 接入 PostgreSQL，初始化 `shares/files/notes/users/folders` 等表。

### 典型问题
- Go 依赖下载超时（`proxy.golang.org i/o timeout`）。
- PostgreSQL 连接失败（密码/端口不一致）。
- Docker daemon 未启动、本地端口冲突（`address already in use`）。

### 解决方式
- 切换 Go 代理：`GOPROXY=https://goproxy.cn,direct`。
- 统一 `DATABASE_URL` 与 Docker 映射端口（常用 `15432`）。
- 先检查端口占用，再启动容器和服务。

## 三、分享功能与文件能力迭代
### 做了什么
- 分享接口可创建 token。
- 分享详情接口可公开查看。
- 分享下载接口可公开下载。
- 文件接口完善：上传、列表、下载、删除、移动。

### 典型问题
- 初始分享下载返回 `NOT_IMPLEMENTED`。
- 早期仅返回 metadata，用户打开是 JSON。

### 解决方式
- 后端补齐分享下载处理。
- 前端分享链接默认切换到 `/download`。

## 四、部署路径演进（踩坑与收敛）
### 早期路径
- 曾尝试“服务器上直接构建 + Docker 运行全部服务”。

### 典型问题
- ECS 规格小，服务器构建不稳定或耗时高。
- Docker Hub 拉取超时。
- Nginx 与容器端口冲突（`bind 0.0.0.0:80 failed`）。

### 收敛后的稳定路径
- 本地编译 `transfer-server`（Linux amd64）。
- `scp` 上传服务器，`systemd` 托管运行。
- Nginx 负责：
  - `/api` 反代到 `127.0.0.1:8080`
  - `/` 提供前端静态文件

## 五、systemd 与 Nginx 上线
### 做了什么
- 新增并启用 `transfer-server.service`。
- 前端 `dist` 上传到 `/opt/transfer/frontend/dist`。
- Nginx 配置前端静态托管 + API 反代。

### 典型问题
- `bind :8080 address already in use`（旧进程未停）。
- `404/401` 交替出现（请求打到旧服务或 token 失效）。
- `413 Request Entity Too Large`（上传大文件被 Nginx 拒绝）。

### 解决方式
- 停掉旧进程后由 systemd 单实例托管。
- 前端代理改为稳定读取 `.env.local`。
- 新增前端 401 自动清 token 机制。
- Nginx 增加 `client_max_body_size`。

## 六、前端开发与体验优化
### 做了什么
- 将设计稿重构为 React 页面，并接入真实 API。
- 增加登录/注册错误提示框。
- 文件卡片按钮重叠问题修复。
- 分享链接复制容错：`Clipboard API` 失败时自动降级复制。
- Office 预览增加超时与失败提示。
- 目录功能接入：目录树、新建文件夹、文件移动。
- 空间占用显示（2GB 配额逻辑）。
- 品牌水印：`Mono's CloudSpace`。

### 典型问题
- 右侧 notes 区域不滚动、内容把页面撑高。
- 预览长时间无反馈。

### 解决方式
- 补全 `h-screen / min-h-0 / h-full / overflow-y-auto` 高度链路。
- 增加预览加载态与超时提示。

## 七、安全与仓库卫生
### 做了什么
- `.gitignore` 清理并补齐：
  - 密钥与本地环境文件
  - 编译产物与缓存
  - 前端打包压缩包
- 增加脚本：
  - `backend/scripts/reset_local_db.sh`
  - `backend/scripts/build_linux_amd64.sh`

### 已执行
- 本地数据库已通过脚本清空，便于全新测试和发布。

## 八、最终稳定流程（建议固定）
### 后端发布
1. 本地编译：
`./backend/scripts/build_linux_amd64.sh`
2. 上传二进制：
`scp ./backend/transfer-server emono:~/transfer-server`
3. 服务器替换并重启：
`sudo mv ~/transfer-server /opt/transfer/backend/transfer-server`
`sudo chmod +x /opt/transfer/backend/transfer-server`
`sudo systemctl restart transfer-server`

### 前端发布
1. 本地打包：
`npm run build`
`COPYFILE_DISABLE=1 tar -czf dist.tar.gz dist`
2. 上传并解压：
`scp dist.tar.gz emono:~/dist.tar.gz`
`sudo tar -xzf ~/dist.tar.gz -C /opt/transfer/frontend`
3. 重载 Nginx：
`sudo nginx -t && sudo systemctl reload nginx`

## 九、这次关键经验
- 小规格服务器不适合频繁在线编译，本地构建更稳定。
- OpenAPI 先行显著降低前后端返工。
- 线上 401/404 很多不是代码错，而是“请求打错目标服务”。
- Nginx 是上线稳定性的关键，建议统一管理上传限制与反代规则。
- 运维脚本化是必须的，能显著减少重复错误。

## 十、下一步建议
- 接入 HTTPS（域名 + 证书）后再启用 Office 预览作为主路径。
- 增加后端“总空间占用”单接口，替代前端多次遍历统计。
- 增加分享有效期、一次性链接、操作审计日志。
- 前端继续拆分组件（`App.jsx` 继续模块化）提升长期可维护性。
