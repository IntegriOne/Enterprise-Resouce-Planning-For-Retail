# syntax=docker/dockerfile:1.7
# Ubuntu 26.10 for every application stage; Node.js latest 26.x from nodejs.org.
# The official per-release SHA256 manifest is checked before extracting Node.
ARG UBUNTU_VERSION=26.10
ARG NODE_MAJOR=26

FROM ubuntu:${UBUNTU_VERSION} AS node-toolchain
ARG UBUNTU_VERSION
ARG NODE_MAJOR
RUN set -eu; . /etc/os-release; test "$ID" = "ubuntu"; test "$VERSION_ID" = "$UBUNTU_VERSION"
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl xz-utils libstdc++6 libatomic1 \
    && rm -rf /var/lib/apt/lists/*
RUN set -eu; \
    arch="$(dpkg --print-architecture)"; \
    case "$arch" in \
      amd64) node_arch=x64 ;; \
      arm64) node_arch=arm64 ;; \
      *) echo "Unsupported Node.js architecture: $arch" >&2; exit 1 ;; \
    esac; \
    base_url="https://nodejs.org/dist/latest-v${NODE_MAJOR}.x"; \
    curl --fail --show-error --silent --location --retry 3 \
      "$base_url/SHASUMS256.txt" -o /tmp/SHASUMS256.txt; \
    archive="$(awk -v arch="$node_arch" '$2 ~ "-linux-" arch "\\.tar\\.xz$" {print $2; exit}' /tmp/SHASUMS256.txt)"; \
    test -n "$archive"; \
    curl --fail --show-error --silent --location --retry 3 \
      "$base_url/$archive" -o "/tmp/$archive"; \
    awk -v wanted="$archive" '$2 == wanted {print; found=1} END {if (!found) exit 1}' \
      /tmp/SHASUMS256.txt > /tmp/node.sha256; \
    (cd /tmp && sha256sum --check node.sha256); \
    tar -xJf "/tmp/$archive" -C /usr/local --strip-components=1; \
    rm -f "/tmp/$archive" /tmp/SHASUMS256.txt /tmp/node.sha256; \
    test "$(node -p 'process.versions.node.split(".")[0]')" = "$NODE_MAJOR"; \
    node --version; npm --version

FROM node-toolchain AS dependencies
WORKDIR /app
COPY package*.json ./
# Uses npm ci deterministically with committed frontend/package-lock.json.
RUN if [ -f package-lock.json ]; then \
      npm ci --no-audit --no-fund; \
    else \
      npm install --no-audit --no-fund; \
    fi

FROM node-toolchain AS build
WORKDIR /app
COPY --from=dependencies /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM ubuntu:${UBUNTU_VERSION} AS runtime
ARG UBUNTU_VERSION
ARG NODE_MAJOR
RUN set -eu; . /etc/os-release; test "$ID" = "ubuntu"; test "$VERSION_ID" = "$UBUNTU_VERSION"
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates libstdc++6 libatomic1 \
    && rm -rf /var/lib/apt/lists/*
COPY --from=node-toolchain /usr/local/ /usr/local/
RUN test "$(node -p 'process.versions.node.split(".")[0]')" = "$NODE_MAJOR" && node --version
ENV NODE_ENV=production
WORKDIR /app
COPY --chown=10001:10001 --from=build /app/.next ./.next
COPY --chown=10001:10001 --from=build /app/public ./public
COPY --chown=10001:10001 --from=build /app/node_modules ./node_modules
COPY --chown=10001:10001 --from=build /app/package.json ./package.json
COPY --chown=10001:10001 --from=build /app/next.config.ts ./next.config.ts
USER 10001:10001
EXPOSE 3000
CMD ["node", "node_modules/next/dist/bin/next", "start", "-p", "3000", "-H", "0.0.0.0"]
