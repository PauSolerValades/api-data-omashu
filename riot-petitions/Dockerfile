# ---------------- BUILD STAGE ----------------------
ARG PLATFORM
ARG TARGETARCH
ARG USER_ID=1001
ARG GROUP_ID=1001

FROM --platform=$PLATFORM golang:1.22.4-alpine AS builder

WORKDIR /usr/src/app

ARG GITHUB_TOKEN

# Install dependencies for Kafka
RUN apk add --no-cache gcc g++ librdkafka-dev musl-dev pkgconf git

# set up to insert custom GO libraries not published
RUN git config --global url."https://oauth2:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
ENV GOPRIVATE=github.com/Omashu-Data
ENV GITHUB_TOKEN=${GITHUB_TOKEN}

COPY ./go.mod ./go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build the Go application with CGO enabled and musl tag
RUN CGO_ENABLED=1 GOOS=$PLATFORM go build -tags musl -o /app/main


# ------------- RUN STAGE -----------------
FROM --platform=$PLATFORM alpine:latest

ARG USER_ID=1001
ARG GROUP_ID=1001

# Install ca-certificates and librdkafka
RUN apk --no-cache add ca-certificates librdkafka

# Create a non-root user with the specified UID and GID
RUN addgroup -g $GROUP_ID gogroup && \
    adduser -D -u $USER_ID -G gogroup -s /bin/sh gouser

# Set environment variable for log directory
ENV LOG_DIR=/app/logs

# Create necessary directories and set permissions
RUN mkdir -p $LOG_DIR && \
    chown -R gouser:gogroup /app

# Copy the built Go application
COPY --from=builder /app/main /app/

# Ensure the binary is executable and set ownership
RUN chmod +x /app/main && \
    chown gouser:gogroup /app/main

# Switch to the non-root user
USER gouser

# Set the working directory
WORKDIR /app

# Run the application
CMD ["./main"]
