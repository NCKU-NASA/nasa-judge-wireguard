FROM golang:1.20 as builder

RUN apt-get update && apt-get full-upgrade -y && apt-get install make -y

WORKDIR /src
COPY go.mod .
COPY go.sum .
ARG CACHE
RUN --mount=type=cache,target="$CACHE" go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get
COPY . .
RUN --mount=type=cache,target="$CACHE" make clean && make

FROM debian:latest as release

RUN apt-get update && apt-get full-upgrade -y && apt-get install ca-certificates curl wget wireguard jq iptables iproute2 iputils-ping dnsutils -y && update-ca-certificates

COPY --from=builder /src/bin /app
COPY docker-entrypoint.sh /app

RUN chmod +x /app/docker-entrypoint.sh && \
    mkdir /config

WORKDIR /app

RUN touch .env
RUN touch init.sh

ENTRYPOINT ["./docker-entrypoint.sh"]
