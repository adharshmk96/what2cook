FROM oven/bun:1.3.13 AS ui

WORKDIR /ui
COPY what2cook-ui/package.json what2cook-ui/bun.lock ./
RUN bun install --no-save
COPY what2cook-ui/ ./
RUN bun run build

FROM golang:1.26-bookworm AS build

WORKDIR /src
COPY what2cook-api/go.mod what2cook-api/go.sum ./
RUN go mod download
COPY what2cook-api/ ./
RUN rm -rf web/dist
COPY --from=ui /ui/dist/ ./web/dist/
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/what2cook .

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir /data \
    && chown 65532:65532 /data
COPY --from=build /out/what2cook /usr/local/bin/what2cook

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["what2cook", "serve"]
