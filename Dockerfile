
# Stage 1: Build golang dependencies and binaries
FROM ubuntu:24.10 AS build
RUN apt-get -qq update && \
    apt-get -qq install golang -y ca-certificates

    WORKDIR /app
    
    # Copy the Go module files
    COPY go.mod .
    COPY go.sum .
    
    # Download the Go module dependencies
    RUN go mod download
    
    COPY . .
    
    RUN go build -o /websocket
     
    FROM alpine:latest as run
    
    # Copy the application executable from the build image
  


# Stage 2: Run Juno with the plugin
FROM ubuntu:24.10

# Install necessary runtime dependencies
RUN apt-get -qq update && \
    apt-get -qq install -y  

WORKDIR /app

# Copy the Juno binary and the plugin from the build stage
COPY --from=build /websocket /websocket
COPY .env ./

EXPOSE 8080
CMD ["/websocket"]
# Run Juno with the plugin
