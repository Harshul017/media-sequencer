# ---------- Build stage ----------
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o media-sequencer .

# ---------- Final stage ----------
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/media-sequencer .

EXPOSE 8080

CMD ["./media-sequencer"]
