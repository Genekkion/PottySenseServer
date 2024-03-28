# Use the official Golang image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /go/src/app

# Copy go mod and sum files
COPY ./telebot/go.mod ./telebot/go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY ./telebot .

# Build the Go app
RUN go build -o main .

# Command to run the executable
CMD ["./main"]
