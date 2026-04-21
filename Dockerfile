FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o envsetup ./cmd/envsetup/

FROM alpine:3.19

RUN apk add --no-cache \
    bash \
    curl \
    wget \
    git \
    zsh \
    sudo \
    docker \
    && rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /build/envsetup /app/envsetup
COPY config/ ./config/

RUN chmod +x /app/envsetup

ENV ENVSETUP_CONFIG=/app/config/tools.yaml

CMD ["/app/envsetup"]