# syntax=docker/dockerfile:1

ARG GO_VERSION=1.27.4
ARG TARGETARCH=amd64

FROM ubuntu:26.10 AS builder

ARG GO_VERSION
ARG TARGETARCH

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl tar \
    && case "${TARGETARCH}" in \
        amd64) GO_SHA256="7226736c2339719cda28fbf3a54d71aa63bcefe376e98cef00034268dab24840" ;; \
        arm64) GO_SHA256="3392582988a273a87ab5043ee1b2c9c17a2c3c00580bf9326c47a2e0aaaeac92" ;; \
        *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" -o /tmp/go.tgz \
    && echo "${GO_SHA256}  /tmp/go.tgz" | sha256sum -c - \
    && tar -C /usr/local -xzf /tmp/go.tgz \
    && rm -f /tmp/go.tgz \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /workspace/backend

# Salin modul shared & enterprise_foundation (build context dari direktori backend/).
COPY shared/authcontext/go.mod ./shared/authcontext/
COPY shared/authcontext ./shared/authcontext/
COPY services/enterprise_foundation/go.mod services/enterprise_foundation/go.sum ./services/enterprise_foundation/

WORKDIR /workspace/backend/services/enterprise_foundation
RUN go mod download

WORKDIR /workspace/backend
COPY services/enterprise_foundation ./services/enterprise_foundation/

WORKDIR /workspace/backend/services/enterprise_foundation
RUN CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH}" go build -trimpath -ldflags="-s -w" -o /out/enterprise-foundation ./cmd \
    && CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH}" go build -trimpath -ldflags="-s -w" -o /out/grpc-healthcheck ./cmd/healthcheck \
    && chmod 0755 /out/enterprise-foundation /out/grpc-healthcheck

FROM ubuntu:26.10 AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid 10001 --system appgroup \
    && useradd --uid 10001 --gid 10001 --system --no-create-home --shell /usr/sbin/nologin appuser

ENV TZ=Asia/Jakarta \
    APP_ENV=production \
    EF_SERVICE_NAME=integrione-enterprise-foundation \
    EF_GRPC_HOST=0.0.0.0 \
    EF_GRPC_PORT=9002

COPY --from=builder /out/enterprise-foundation /usr/local/bin/enterprise-foundation
COPY --from=builder /out/grpc-healthcheck /usr/local/bin/grpc-healthcheck

RUN ln -snf /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && echo "Asia/Jakarta" > /etc/timezone

USER 10001:10001
EXPOSE 9002

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/grpc-healthcheck"]

ENTRYPOINT ["/usr/local/bin/enterprise-foundation"]
