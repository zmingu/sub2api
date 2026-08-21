# Custom theme image deployment

## Local development

The custom Compose file uses the locally built image by default:

```powershell
docker build -f deploy\Dockerfile -t sub2api-custom:local .
docker compose -f deploy\docker-compose.custom.yml up -d
```

## GHCR deployment

Set a fixed public image tag in `deploy/.env` (this file is ignored by Git):

```dotenv
SUB2API_IMAGE=ghcr.io/zmingu/sub2api-custom:0.1.177-theme2
```

Then pull and recreate only the application image:

```powershell
docker compose -f deploy\docker-compose.custom.yml pull sub2api
docker compose -f deploy\docker-compose.custom.yml up -d sub2api
```

PostgreSQL and Redis data remain in `deploy/postgres_data` and
`deploy/redis_data`; replacing the application image does not remove them.

## Architectures

`.github/workflows/custom-image.yml` publishes a manifest list covering
`linux/amd64` and `linux/arm64`, so the same tag works on x86 servers and on
ARM boxes (Ampere, Graviton, Raspberry Pi 5, Apple Silicon) with no change to
the Compose file — `postgres:18-alpine` and `redis:8-alpine` ship arm64 too.

Verify a published tag carries both platforms:

```bash
docker buildx imagetools inspect ghcr.io/zmingu/sub2api-custom:0.1.177-theme4
```

A local `docker build` only produces a single-arch image for the host. To
produce a multi-arch image by hand, use buildx and push directly (buildx
cannot `--load` a manifest list into the local image store):

```bash
docker buildx build --platform linux/amd64,linux/arm64   -f deploy/Dockerfile -t ghcr.io/zmingu/sub2api-custom:test --push .
```

When moving an existing deployment from x86 to ARM, do not copy
`deploy/postgres_data` across architectures — dump and restore instead. See
`deploy/CUSTOM_THEME_WORKFLOW.md`.

## Rollback

Change `SUB2API_IMAGE` to the matching upstream image in `deploy/.env`:

```dotenv
SUB2API_IMAGE=weishaw/sub2api:0.1.177
```

Then run the same `pull` and `up` commands. Keep the current `deploy/data`,
`deploy/postgres_data`, and `deploy/redis_data` directories intact.
