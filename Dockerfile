FROM golang:1.24-bookworm AS builder

# Set the working directory
WORKDIR /app

# Copy and download dependencies
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY . /app

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o ddns .

FROM debian:bookworm-slim

RUN groupadd -r app && useradd -r -d /app -g app -N app

COPY --from=builder --chown=app:app /app/ddns /app/ddns
USER app

WORKDIR /app


ENTRYPOINT [ "/app/ddns" ]