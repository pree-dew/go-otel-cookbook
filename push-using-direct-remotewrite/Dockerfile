# Build docker image for my main.go file

# Use the official Golang image to create a build artifact.

FROM golang:1.21 as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.mod go.sum ./

RUN go mod download

# Copy local code to the container image.
COPY . .

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o server

# Use the official Alpine image for a lean production container.
# https://hub.docker.com/_/alpine
# https://docs.docker.com/samples/golang/
FROM alpine:3.12

# Expose the port the app runs on
EXPOSE 8081

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /server

ENTRYPOINT ["/server"]



