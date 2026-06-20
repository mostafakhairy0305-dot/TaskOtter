FROM golang:1.26.4-alpine3.22 AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /taskotter ./cmd/taskotter-sync

FROM alpine:3.22

RUN apk add --no-cache ca-certificates git

COPY --from=builder /taskotter /taskotter

ENTRYPOINT ["/taskotter"]
