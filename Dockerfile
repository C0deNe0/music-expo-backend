# ---- Build Stage ----
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o api ./cmd/api
RUN go build -o worker ./cmd/worker

# ---- Runtime Stage ----
FROM ubuntu:22.04

WORKDIR /app

# install dependencies
RUN apt-get update && \
    apt-get install -y \
    ca-certificates \
    curl \
    python3 \
    ffmpeg \
    nodejs \
    npm && \
    rm -rf /var/lib/apt/lists/*

# install yt-dlp
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
    -o /usr/local/bin/yt-dlp && \
    chmod +x /usr/local/bin/yt-dlp

# copy binaries
COPY --from=builder /app/api .
COPY --from=builder /app/worker .

# expose port
EXPOSE 8080

# run both api + worker
CMD ["sh", "-c", "./worker & ./api"]