# Ubuntu 26 application runtime, using the latest official Caddy executable.
FROM ubuntu:26.10
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=caddy:latest /usr/bin/caddy /usr/bin/caddy
WORKDIR /etc/caddy
EXPOSE 443
CMD ["/usr/bin/caddy", "run"]
