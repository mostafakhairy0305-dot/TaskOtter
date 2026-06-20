FROM golang:1.23-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /taskotter-sync ./cmd/taskotter-sync

FROM alpine:3.20

RUN apk add --no-cache ca-certificates git && \
    addgroup -S taskotter && adduser -S taskotter -G taskotter

COPY --from=builder /taskotter-sync /taskotter-sync

USER taskotter
ENTRYPOINT ["/taskotter-sync"]
