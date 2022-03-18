#
# Build Go binaries
#
FROM golang:1.17 as builder

# Define the build path
WORKDIR /root/go/src/github.com/seatgeek/hashi-helper

# Disable CGO
ENV CGO_ENABLED=0

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o /bin/hashi-helper

#
# Runtime container
#
FROM ubuntu:bionic

# Copy the binary from builder
COPY --from=builder /bin/hashi-helper /hashi-helper

# Configure entrypoint
ENTRYPOINT ["/hashi-helper"]
