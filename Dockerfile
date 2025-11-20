FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /app/pr-manager ./cmd/app/main.go

FROM alpine:latest AS final

WORKDIR /app

COPY --from=builder /app/pr-manager ./pr-manager
COPY .env ./

EXPOSE 8080
CMD ["./pr-manager"]