# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod & go.sum lalu download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Copy .env file for runtime environment variables
COPY .env .

# Build binary (nama binary = app)
RUN go build -o main main.go

# Stage 2: Run
FROM alpine:3.19

WORKDIR /root/

# Copy binary dari builder
COPY --from=builder /app/main .

# Copy .env file from build context to runtime container
COPY .env .

RUN ls -lah

RUN chmod +x ./main

# Set default command
CMD ["./main"]