
# Use official Golang image (Go 1.21) to build the app
FROM golang

# Set the working directory
WORKDIR /app

# Copy the Go source code into the container
COPY . .

# Download dependencies and build the Go app
RUN go mod tidy
RUN go build -o websocket-server .

# Use a smaller base image to run the app

# Set the working directory to where the Go binary is
WORKDIR /app

# Expose port 4296 for the WebSocket server
EXPOSE 4296

# Run the WebSocket server directly from the working directory
CMD  ["./websocket-server"]





