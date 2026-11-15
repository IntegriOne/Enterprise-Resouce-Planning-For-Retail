# syntax=docker/dockerfile:1

ARG UBUNTU_IMAGE=ubuntu:26.10
ARG GO_VERSION=1.27.4

# Stage 1: Menyiapkan toolchain Go yang terverifikasi checksum resminya.
FROM ${UBUNTU_IMAGE} AS go-toolchain

ARG GO_VERSION
ARG TARGETARCH

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        curl && \
    rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    arch="${TARGETARCH:-amd64}"; \
    case "${arch}" in \
        amd64) sha="7226736c2339719cda28fbf3a54d71aa63bcefe376e98cef00034268dab24840" ;; \
        arm64) sha="3392582988a273a87ab5043ee1b2c9c17a2c3c00580bf9326c47a2e0aaaeac92" ;; \
        *) echo "unsupported architecture: ${arch}" >&2; exit 1 ;; \
    esac; \
    curl -fsSL "https://dl.google.com/go/go${GO_VERSION}.linux-${arch}.tar.gz" -o /tmp/go.tar.gz; \
    echo "${sha}  /tmp/go.tar.gz" | sha256sum -c -; \
    tar -C /usr/local -xzf /tmp/go.tar.gz; \
    rm -f /tmp/go.tar.gz

# Stage 2: Mengunduh dependency module gateway, shared, dan kontrak IAM.
FROM go-toolchain AS deps

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /src
COPY shared /src/shared
COPY services/iam/go.mod services/iam/go.sum /src/services/iam/
COPY gateway/go.mod gateway/go.sum /src/gateway/

WORKDIR /src/gateway
RUN go mod download

# Stage 3: Mengompilasi binary statis gateway-server dan gateway-healthcheck.
FROM deps AS builder

COPY services/iam /src/services/iam
COPY gateway /src/gateway

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway-server ./cmd && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway-healthcheck ./cmd/healthcheck

# Stage 4: Runtime container minimal Ubuntu 26.10 dengan user non-root UID/GID 10001.
FROM ${UBUNTU_IMAGE} AS runtime

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        tzdata && \
    ln -snf /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && \
    echo "Asia/Jakarta" > /etc/timezone && \
    rm -rf /var/lib/apt/lists/* && \
    if ! getent group 10001 >/dev/null; then groupadd --gid 10001 app; fi && \
    if ! getent passwd 10001 >/dev/null; then useradd --uid 10001 --gid 10001 --create-home --shell /usr/sbin/nologin app; fi

ENV TZ=Asia/Jakarta \
    APP_ENV=development \
    GATEWAY_SERVICE_NAME=integrione-gateway \
    GATEWAY_HTTP_HOST=0.0.0.0 \
    GATEWAY_HTTP_PORT=8080

WORKDIR /app

COPY --from=builder /out/gateway-server /usr/local/bin/gateway-server
COPY --from=builder /out/gateway-healthcheck /usr/local/bin/gateway-healthcheck

RUN chmod 0755 /usr/local/bin/gateway-server /usr/local/bin/gateway-healthcheck

USER 10001:10001

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=5 \
    CMD ["/usr/local/bin/gateway-healthcheck"]

ENTRYPOINT ["/usr/local/bin/gateway-server"]
