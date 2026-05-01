FROM golang:1.24-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o vitualizz-devstack ./cmd/vitualizz-devstack/

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

COPY --from=builder /build/vitualizz-devstack /app/vitualizz-devstack

RUN chmod +x /app/vitualizz-devstack

CMD ["/app/vitualizz-devstack"]