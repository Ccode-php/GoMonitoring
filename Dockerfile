FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o scanner .

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y \
    iputils-ping \
    net-tools \
    snmp \
    snmpd \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/scanner /scanner

CMD ["/scanner"]