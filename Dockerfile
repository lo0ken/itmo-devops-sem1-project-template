FROM golang:1.23.3-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o app .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/app .

EXPOSE 8080

CMD ["./app"]