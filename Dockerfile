# Using go 1.21 as the base image
FROM golang:1.21

# Declare a working directory inside the container
WORKDIR /usr/src/app

# Copy all the project files into the container's "/usr/src/app" directory
COPY . /usr/src/app/

# Run the program's build commands and outputting the result to "app"
RUN go mod download && go mod verify && go build -o app

# Command to run the program
CMD ["./app"]