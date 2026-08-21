# Sub2API 自定义主题完整工作流

本文记录 Fork 仓库的自定义主题、Docker、GHCR、远程无损更新和上游同步流程。

## 当前项目约定

| 项目 | 值 |
| --- | --- |
| Fork | `https://github.com/zmingu/sub2api.git` |
| 上游 | `https://github.com/Wei-Shaw/sub2api.git` |
| 本地项目目录 | `F:\sub2api\sub2api` |
| 远程部署目录 | `/opt/stacks/sub2api-deploy` |
| 自定义 Compose | `deploy/docker-compose.custom.yml` |
| 自定义工作流 | `.github/workflows/custom-image.yml` |
| GHCR 镜像 | `ghcr.io/zmingu/sub2api-custom` |

## 重要数据安全规则

远程更新只替换 `sub2api` 应用容器，不删除数据库和 Redis 数据。

**禁止在生产环境执行：**

```bash
docker compose down -v
rm -rf data postgres_data redis_data
docker volume prune
```

必须保留这些目录：

```text
data
postgres_data
redis_data
```

远程部署前应备份：

```text
docker-compose.yml
docker-compose.custom.yml
.env
```

## 一、首次本地准备

### 1. 配置 Git 上游

```powershell
cd F:\sub2api\sub2api
git remote add upstream https://github.com/Wei-Shaw/sub2api.git
git remote -v
```

如果 `upstream` 已存在，不要重复添加。

### 2. Windows 环境

本地构建需要：

- Docker Desktop
- WSL 2 后端
- Node.js
- Corepack
- pnpm `9.15.9`

启用 pnpm：

```powershell
corepack enable
corepack prepare pnpm@9.15.9 --activate
pnpm --version
```

项目 Dockerfile 也固定使用 pnpm `9.15.9`，本地和 CI 保持一致。

## 二、修改主题

常用文件：

```text
frontend/src/style.css              全局样式、按钮、卡片、字体、背景
frontend/tailwind.config.js         Tailwind 颜色、字体和阴影 token
frontend/public/fonts/              Maple Mono NF CN 字体文件及许可证
frontend/src/views/HomeView.vue     首页结构和首页样式
frontend/src/components/**/*.vue    局部公共组件或页面样式
frontend/public/wallpapers/         本地壁纸资源
```

品牌内容优先使用后台设置：

- 站点名称
- Logo
- Favicon

不要在组件中硬编码这些值。首页使用 `appStore` 的公共设置，后台修改后即可生效。

## 三、本地验证

### 1. 前端构建

```powershell
cd F:\sub2api\sub2api\frontend
corepack pnpm install --frozen-lockfile
corepack pnpm run build
```

构建产物会写入：

```text
backend/internal/web/dist/
```

### 字体

全站使用 Maple Mono NF CN。字体文件随前端镜像打包，定义在 `frontend/src/style.css`，因此远程服务器不依赖本机安装字体。当前使用 WOFF2 中文子集，保留项目源码中实际使用的字符；后台动态内容遇到未包含字符时会回退到系统等宽字体。

字体文件和 `LICENSE.txt` 位于：

```text
frontend/public/fonts/
```

### 2. Docker 构建

```powershell
cd F:\sub2api\sub2api
docker build --pull -f deploy\Dockerfile -t sub2api-custom:local .
```

如果出现 Go 版本错误，确认 `deploy/Dockerfile` 的 Go 镜像不低于 `backend/go.mod` 要求的版本。

如果出现 Node heap out of memory，确认 Dockerfile 的 `NODE_OPTIONS` 有足够堆内存。

本地 `docker build` 只产出**当前主机架构**的单架构镜像，用于本机验证足够。
多架构镜像（amd64 + arm64）由 CI 产出，见第四节。如果要在本地手工产出多架构镜像：

```powershell
docker buildx build --platform linux/amd64,linux/arm64 `
  -f deploy\Dockerfile -t ghcr.io/zmingu/sub2api-custom:test --push .
```

buildx 不支持把多架构镜像 `--load` 进本地 docker images，只能 `--push` 到仓库。

### 3. 本地 Compose

本地 `deploy/.env` 不提交 Git。首次运行需要设置至少：

```dotenv
POSTGRES_PASSWORD=随机强密码
JWT_SECRET=随机强密钥
TOTP_ENCRYPTION_KEY=随机强密钥
```

启动：

```powershell
cd F:\sub2api\sub2api
docker compose -f deploy\docker-compose.custom.yml config
docker compose -f deploy\docker-compose.custom.yml up -d --no-build
```

检查：

```powershell
docker compose -f deploy\docker-compose.custom.yml ps
curl.exe http://localhost:8080/health
```

预期健康检查：

```json
{"status":"ok"}
```

## 四、提交并发布 GHCR

检查：

```powershell
git diff --check
git status
git diff --stat
```

提交并推送：

```powershell
git add .
git commit -m "update frontend theme"
git push origin main
```

`.github/workflows/custom-image.yml` 的规则：

- push 到 `main`：构建并推送镜像
- Pull Request：只构建验证，不推送
- 手动触发：允许重新发布

查看 Actions：

```text
https://github.com/zmingu/sub2api/actions
```

镜像标签示例：

```text
ghcr.io/zmingu/sub2api-custom:0.1.177-theme2
ghcr.io/zmingu/sub2api-custom:latest
```

生产部署优先使用固定版本标签，不要依赖 `latest`。

### 多架构验证

workflow 构建 `linux/amd64,linux/arm64` 两个架构并合成 manifest list。
推送完成后确认两个架构都在：

```bash
docker buildx imagetools inspect ghcr.io/zmingu/sub2api-custom:0.1.177-theme4
```

预期输出包含：

```text
Platform:  linux/amd64
Platform:  linux/arm64
```

只出现一个平台说明 workflow 的 `platforms:` 没生效，ARM 机器会拉不到镜像。

构建时长：arm64 的最终 runtime 层（`apk add` / `adduser`）跑在 QEMU 下，
比单架构构建多几分钟属正常。前端和 Go 编译仍在 runner 原生 amd64 上跑，
不受 QEMU 影响（`deploy/Dockerfile` 的 builder 阶段用 `--platform=$BUILDPLATFORM`
交叉编译，`GOARCH=${TARGETARCH}` 决定输出架构）。

## 五、远程服务器无损更新

远程目录：

```bash
cd /opt/stacks/sub2api-deploy
```

在远程 `.env` 中设置固定镜像：

```dotenv
SUB2API_IMAGE=ghcr.io/zmingu/sub2api-custom:0.1.177-theme2
```

如果使用 `docker-compose.custom.yml`，它支持：

```yaml
image: ${SUB2API_IMAGE:-sub2api-custom:local}
```

部署前检查：

```bash
docker compose ps
docker inspect sub2api sub2api-postgres sub2api-redis \
  --format '{{.Name}} {{range .Mounts}}{{.Source}} -> {{.Destination}}; {{end}}'
```

拉取并只更新应用：

```bash
docker compose pull sub2api
docker compose up -d --no-deps sub2api
```

不要使用全量 `down`，不要重建 PostgreSQL/Redis。

验证：

```bash
docker compose ps
curl -fsS http://127.0.0.1:8080/health
docker inspect sub2api --format '{{.Config.Image}}'
```

预期：

```text
sub2api       healthy
sub2api-postgres healthy
sub2api-redis healthy
```

如果使用外部反向代理或域名，再从浏览器访问实际域名检查页面。

### ARM 机器（aarch64）部署

镜像是多架构 manifest list，ARM 机器上的命令与 x86 完全一致，
docker 会自动选中 `linux/arm64` 变体。无需修改 `docker-compose.custom.yml`：
依赖的 `postgres:18-alpine`、`redis:8-alpine` 官方均提供 arm64。

首次部署前置检查：

```bash
uname -m                      # 预期 aarch64
docker login ghcr.io          # GHCR package 若为 private 必须先登录
docker compose pull sub2api
docker inspect sub2api --format '{{.Architecture}}'   # 预期 arm64
```

GHCR package 首次推送默认是 private。若不想在服务器上存登录凭据，
到 `https://github.com/users/zmingu/packages/container/sub2api-custom/settings`
把可见性改为 public。

### 从 x86 机器迁移到 ARM 机器

**不要直接复制 `deploy/postgres_data`。** 那是 bind mount 的 PGDATA 原始数据目录，
跨 CPU 架构复制 PGDATA，PostgreSQL 官方不保证兼容。正确做法是逻辑迁移：

```bash
# 旧机器（x86）导出
docker compose exec -T postgres pg_dump -U sub2api -Fc sub2api > sub2api.dump

# 新机器（ARM）：先起 postgres，再恢复
docker compose -f deploy/docker-compose.custom.yml up -d postgres
docker compose -f deploy/docker-compose.custom.yml exec -T postgres   pg_restore -U sub2api -d sub2api --clean --if-exists < sub2api.dump

# 最后起应用
docker compose -f deploy/docker-compose.custom.yml up -d
```

`deploy/redis_data` 的 RDB/AOF 是架构无关的，可以直接复制。
`deploy/.env` 里的 `JWT_SECRET`、`TOTP_ENCRYPTION_KEY` 必须原样带到新机器，
否则已签发的 token 和已绑定的 TOTP 全部失效。

## 六、回滚

### 回滚到上一版自定义镜像

```dotenv
SUB2API_IMAGE=ghcr.io/zmingu/sub2api-custom:0.1.177-theme1
```

```bash
docker compose pull sub2api
docker compose up -d --no-deps sub2api
```

### 回滚到官方同版本镜像

```dotenv
SUB2API_IMAGE=weishaw/sub2api:0.1.177
```

```bash
docker compose pull sub2api
docker compose up -d --no-deps sub2api
```

回滚时也不要删除：

```text
data
postgres_data
redis_data
```

## 七、同步上游

先确认工作区干净：

```powershell
cd F:\sub2api\sub2api
git status
git fetch upstream
```

合并上游：

```powershell
git checkout main
git merge upstream/main
```

重点检查可能冲突的文件：

```text
frontend/src/style.css
frontend/tailwind.config.js
frontend/src/views/HomeView.vue
backend/internal/server/routes/common.go
deploy/Dockerfile
.github/workflows/custom-image.yml
```

解决冲突后验证：

```powershell
git add .
corepack pnpm --dir frontend run build
git diff --check
git commit -m "sync upstream and preserve custom theme"
git push origin main
```

Actions 成功后，再到远程服务器更新固定版本镜像。

## 八、日常最短流程

```powershell
cd F:\sub2api\sub2api
# 修改 frontend/ 下的主题文件
corepack pnpm --dir frontend run build
git diff --check
git add .
git commit -m "update frontend theme"
git push origin main
```

然后在远程服务器：

```bash
cd /opt/stacks/sub2api-deploy
# 修改 .env 中的 SUB2API_IMAGE 为新固定标签
docker compose pull sub2api
docker compose up -d --no-deps sub2api
docker compose ps
curl -fsS http://127.0.0.1:8080/health
```
