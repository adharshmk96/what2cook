# what2cook-api

Go API with the Vue UI embedded at `/` (SPA routes under `/app`).

## Build (from repo root)

```bash
make build
```

Produces `./what2cook` in this directory.

## Run

```bash
cp config.yaml.example config.yaml
./what2cook serve
```

- `GET /healthz`
- ` /api/v1/auth/*` — auth API
- `/` — landing (embedded SPA)
- `/app/*` — SPA fallback for login, dashboard, etc.

Config: `config.yaml` or env `WHAT2COOK_*` (see `config.yaml.example`).
