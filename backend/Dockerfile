FROM golang:1.23-alpine

WORKDIR /app

# Install required system dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Use a minimal alpine image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=0 /app/main .
COPY --from=0 /app/.env .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"]
