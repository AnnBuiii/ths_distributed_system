FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY fastdb.go doc.go ./
COPY event/ ./event/
COPY persist/ ./persist/
COPY server/ ./server/

WORKDIR /app/server
RUN go build -o /app/server-binary .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server-binary ./server

EXPOSE 6001 6002 6003 6004

ENTRYPOINT ["./server"]
