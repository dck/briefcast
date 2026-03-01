FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server
RUN CGO_ENABLED=0 go build -o worker ./cmd/worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates ffmpeg
WORKDIR /app
COPY --from=builder /app/server /app/worker ./
COPY --from=builder /app/templates/*.html ./templates/
EXPOSE 8080
CMD ["./server"]
