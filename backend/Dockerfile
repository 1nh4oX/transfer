FROM golang:1.26-alpine AS builder
WORKDIR /src

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/transfer-server ./cmd/server

FROM alpine:3.21
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /out/transfer-server /app/transfer-server
RUN mkdir -p /app/uploads && chown -R app:app /app

USER app
EXPOSE 8080

CMD ["/app/transfer-server"]
