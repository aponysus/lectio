FROM node:18-bookworm AS web-build
WORKDIR /app/web

COPY web/package.json ./
RUN npm install

COPY web/ ./
RUN npm run build

FROM golang:1.25-bookworm AS go-build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-build /app/web/dist ./web/dist

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/lectio ./cmd/lectio

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=go-build /out/lectio /usr/local/bin/lectio
COPY --from=go-build /app/migrations ./migrations
COPY --from=go-build /app/web/dist ./web/dist

ENV LECTIO_ADDR=:8080
ENV LECTIO_DB_PATH=/data/lectio.db
ENV LECTIO_WEB_DIST=/app/web/dist

EXPOSE 8080

CMD ["lectio", "serve"]
