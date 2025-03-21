FROM golang:alpine AS builder

WORKDIR /build

RUN mkdir /build-cache
RUN go env -w GOCACHE=/build-cache

RUN apk add go-task

ADD go.mod go.sum .

RUN go mod download

COPY . .

RUN --mount=type=cache,target=/build-cache go-task build

FROM alpine

WORKDIR /app

COPY --from=builder /build/bin/* /app

CMD ["./torr-helper_amd64", "config.yml"]
