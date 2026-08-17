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

### 2. Docker 构建

```powershell
cd F:\sub2api\sub2api
docker build --pull -f deploy\Dockerfile -t sub2api-custom:local .
```

如果出现 Go 版本错误，确认 `deploy/Dockerfile` 的 Go 镜像不低于 `backend/go.mod` 要求的版本。

如果出现 Node heap out of memory，确认 Dockerfile 的 `NODE_OPTIONS` 有足够堆内存。

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
