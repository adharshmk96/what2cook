# what2cook

Single-binary recipe app: Go API (`what2cook-api`) + Vue UI (`what2cook-ui`) embedded at `/` (app routes under `/app`).

## Prerequisites

- [Go](https://go.dev/dl/) 1.22+
- [Bun](https://bun.sh/) (UI build)
- [Task](https://taskfile.dev/) (`task`) for ops scripts

## Setup

From the repo root (requires Go + Bun):

```bash
task init
```

This installs Go modules and UI deps, and creates `what2cook-api/config.yaml` from the example if missing.

## Build

From the repo root:

```bash
task build
```

Or `make build`. This:

1. Builds the UI (`bun run build` → `what2cook-ui/dist`)
2. Copies it into `what2cook-api/web/dist` (Go `embed`)
3. Compiles `what2cook-api/what2cook`

## Run

```bash
cp what2cook-api/config.yaml.example what2cook-api/config.yaml
# edit token_secret (and SMTP if you want real email)
cd what2cook-api && ./what2cook serve
```

Or in one step after config exists:

```bash
make run
```

- UI: http://localhost:8080/
- API: http://localhost:8080/api/v1/...
- Health: http://localhost:8080/healthz

Forgot-password reset links are logged when SMTP is unset.

## Dev

Stop any previous development server, build the UI and API, then start the
single server binary:

```bash
task run-dev
```

Stop the server:

```bash
task stop-dev
```

- App: http://localhost:8080/
- API: http://localhost:8080/api/v1/...
- Log/PID: `.dev/server.log` and `.dev/server.pid`

## Layout

```
what2cook/
  Taskfile.yml          # task init / build / run-dev / stop-dev
  scripts/              # ops scripts invoked by Task
  Makefile              # make build / make run
  what2cook-api/        # Gin + GORM + embed
  what2cook-ui/         # Vue 3 + Vite (base /; app routes /app/*)
```
