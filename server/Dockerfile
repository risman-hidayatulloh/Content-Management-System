# ===== DEV (hot reload) =====
FROM golang:1.25-alpine AS dev
WORKDIR /app

# install curl (buat unduh biner air)
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*
RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b /usr/local/bin
RUN which air && air -v
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["air", "-c", ".air.toml"]

# ===== PROD (build biner) =====
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/api
RUN test -f /app/config.toml || true


FROM gcr.io/distroless/static-debian12:nonroot AS prod
WORKDIR /app
COPY --from=build /out/server /app/server

COPY --from=build /app/config.toml /app/config.toml
ENV CONFIG_PATH=/app/config.toml
EXPOSE 4000
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
