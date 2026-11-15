# syntax=docker/dockerfile:1

ARG UBUNTU_IMAGE=ubuntu:26.10
ARG GO_VERSION=1.27.4

# Tahap 1: Mengunduh dan memverifikasi integritas toolchain Go resmi.
FROM ${UBUNTU_IMAGE} AS go-toolchain

ARG GO_VERSION
ARG TARGETARCH

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        tar \
        gzip && \
    rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    arch="${TARGETARCH:-$(dpkg --print-architecture)}"; \
    case "${arch}" in \
        amd64) sha="7226736c2339719cda28fbf3a54d71aa63bcefe376e98cef00034268dab24840" ;; \
        arm64) sha="3392582988a273a87ab5043ee1b2c9c17a2c3c00580bf9326c47a2e0aaaeac92" ;; \
        *) echo "unsupported architecture: ${arch}" >&2; exit 1 ;; \
    esac; \
    curl -fsSL "https://dl.google.com/go/go${GO_VERSION}.linux-${arch}.tar.gz" -o /tmp/go.tar.gz; \
    echo "${sha}  /tmp/go.tar.gz" | sha256sum -c -; \
    tar -C /usr/local -xzf /tmp/go.tar.gz; \
    rm -f /tmp/go.tar.gz

# Tahap 2: Mengunduh dependensi Go dengan build context root backend/ agar modul shared/ ikut tersalin.
FROM go-toolchain AS deps

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /src
COPY shared /src/shared
COPY services/iam/go.mod services/iam/go.sum /src/services/iam/
WORKDIR /src/services/iam
RUN go mod download

# Tahap 3: Mengompilasi binary server IAM dan probe healthcheck gRPC statis.
FROM deps AS builder

COPY services/iam /src/services/iam

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/iam-server ./cmd && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/grpc-healthcheck ./cmd/healthcheck

# Tahap 4: Runtime image minimal Ubuntu 26.10 dengan tzdata lokal OS dan user non-root.
FROM ${UBUNTU_IMAGE} AS runtime

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        tzdata && \
    ln -snf /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && \
    echo "Asia/Jakarta" > /etc/timezone && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /etc/integrione/certs && \
    if ! getent group 10001 >/dev/null; then groupadd --gid 10001 app; fi && \
    if ! getent passwd 10001 >/dev/null; then useradd --uid 10001 --gid 10001 --create-home --shell /usr/sbin/nologin app; fi

ENV TZ=Asia/Jakarta \
    APP_ENV=development \
    IAM_SERVICE_NAME=integrione-iam \
    IAM_GRPC_HOST=0.0.0.0 \
    IAM_GRPC_PORT=9001

WORKDIR /app

COPY --from=builder /out/iam-server /usr/local/bin/iam-server
COPY --from=builder /out/grpc-healthcheck /usr/local/bin/grpc-healthcheck

RUN chmod 0755 /usr/local/bin/iam-server /usr/local/bin/grpc-healthcheck

USER 10001:10001

EXPOSE 9001

HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=5 \
    CMD ["/usr/local/bin/grpc-healthcheck"]

ENTRYPOINT ["/usr/local/bin/iam-server"]
