FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .

# Run 'go test' command to execute all tests in the current directory and all of its subdirectories.
# The '-coverprofile=coverage.out' option generates a test coverage report named 'coverage.out'.
RUN go test ./... -coverprofile=coverage.out

RUN go tool cover -func=coverage.out

# Build the Go application
RUN go build -o konstruktor-test .

# Command to run the executable
CMD ["./konstruktor-test"]