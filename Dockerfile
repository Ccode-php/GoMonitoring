FROM golang:1.24 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o scanner .

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y \
    iputils-ping \
    net-tools \
    snmp \
    snmpd

COPY --from=builder /app/scanner /scanner

CMD ["/scanner"]