FROM golang:1.23-alpine3.20 AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates=20260413-r0 git=2.45.4-r0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /taskotter ./cmd/taskotter-sync

FROM alpine:3.20

RUN apk add --no-cache ca-certificates=20260413-r0 git=2.45.4-r0

COPY --from=builder /taskotter /taskotter

ENTRYPOINT ["/taskotter"]
