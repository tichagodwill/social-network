FROM golang:latest

LABEL Authors="aabdulhu, yalsayya, sayedalawi, smahfood"
LABEL Description="Social-Network Backend"
LABEL Version="1.0"

WORKDIR /app

# Install dependencies
RUN apt-get update && apt-get install -y sqlite3 libsqlite3-dev

# Copy the application
COPY . .

# Build the Go application
RUN go build -o main .

# Create directory for database and source code
RUN mkdir -p /app/data /app/src

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"]