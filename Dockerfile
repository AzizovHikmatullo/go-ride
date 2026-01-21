FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o go-ride ./cmd/go-ride

FROM alpine:3.14

WORKDIR /build

COPY --from=builder /build/go-ride /build/go-ride

COPY --from=builder /build/docs /build/docs

COPY --from=builder /build/migrations /build/migrations

COPY .env ./

EXPOSE 8080

ENTRYPOINT ["/build/go-ride"]
