# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /usrsvr ./cmd/usrsvr

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /usrsvr /usr/local/bin/usrsvr
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/usrsvr"]
