FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . ./

WORKDIR /build/web
RUN CGO_ENABLED=0 GOOS=linux go build -o server -ldflags "-w -s" .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/web/server /app/server
COPY --from=builder /build/web/static /app/static

EXPOSE 8080

CMD ["/app/server"]
