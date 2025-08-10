FROM alpine:latest

RUN apk add --no-cache ca-certificates curl jq

ARG MYCELIUM_VERSION=v0.6.1
RUN curl -sSL "https://github.com/threefoldtech/mycelium/releases/download/${MYCELIUM_VERSION}/mycelium-x86_64-unknown-linux-musl.tar.gz" \
    | tar -C /usr/local/bin -xz && \
    chmod +x /usr/local/bin/mycelium

COPY run.sh /run.sh
RUN chmod +x /run.sh

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD pgrep mycelium || exit 1

ENTRYPOINT ["/run.sh"]
