FROM golang:1.22

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .

# Build the Go application
RUN go build -o konstruktor-test .

# Command to run the executable
CMD ["./konstruktor-test"]