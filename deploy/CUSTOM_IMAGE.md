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

## Rollback

Change `SUB2API_IMAGE` to the matching upstream image in `deploy/.env`:

```dotenv
SUB2API_IMAGE=weishaw/sub2api:0.1.177
```

Then run the same `pull` and `up` commands. Keep the current `deploy/data`,
`deploy/postgres_data`, and `deploy/redis_data` directories intact.
