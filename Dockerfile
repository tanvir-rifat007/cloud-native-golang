FROM golang:1-bullseye AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.release=`git rev-parse --short=8 HEAD`'" -o /bin/api ./cmd/api


# DEV IMAGE

FROM debian:bullseye-slim AS dev
WORKDIR /app
COPY --from=builder /bin/api ./
COPY .env .env
COPY public ./public
COPY templates ./templates
CMD ["./api"]



# PRODUCTION IMAGE
FROM gcr.io/distroless/base-debian11 AS prod
WORKDIR /app
COPY --from=builder /bin/api ./
COPY public ./public
COPY templates ./templates
CMD ["./api"]
