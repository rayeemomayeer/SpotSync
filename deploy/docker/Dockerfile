# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /spotsync ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=build /spotsync .

EXPOSE 8080

USER nobody

CMD ["./spotsync"]
