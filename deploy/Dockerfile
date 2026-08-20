FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/controlplane ./cmd/server

FROM alpine:3.20
RUN addgroup -S infra && adduser -S infra -G infra
COPY --from=builder /out/controlplane /usr/local/bin/controlplane
COPY configs /etc/infra/configs
COPY migrations /etc/infra/migrations
USER infra
ENTRYPOINT ["controlplane"]
CMD ["-config=/etc/infra/configs/config.yaml"]
