# Multi-stage build - builder stage
FROM golang:1.25.3-alpine AS builder

# Install required build tools
RUN apk add --no-cache \
    protobuf-dev \
    protoc \
    git \
    make

WORKDIR /moon

# Copy build files
COPY Makefile Makefile

# Initialize environment
RUN make init

# Copy source code
COPY . .

# Build application
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

WORKDIR /moon

# Copy binary
COPY --from=builder /moon/bin/jade_tree /usr/local/bin/jade_tree

# Ensure executable permission
RUN chmod +x /usr/local/bin/jade_tree

# Expose ports (http=8004, grpc=9004)
EXPOSE 8004 9004

# Run application
ENTRYPOINT ["/usr/local/bin/jade_tree"]
CMD ["run", "all"]
