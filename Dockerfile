FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o app .

FROM ubuntu:24.04

# Faqat kerakli paketlar: ping, ip neigh, TLS sertifikatlar.
# (snmp, snmpd, net-tools kerak emas, snmpd esa keraksiz daemon)
RUN apt-get update && apt-get install -y --no-install-recommends \
    iproute2 \
    iputils-ping \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app /usr/local/bin/app

CMD ["/usr/local/bin/app"]
