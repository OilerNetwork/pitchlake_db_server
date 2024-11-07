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

FROM ubuntu:24.10

# Install necessary runtime dependencies
RUN apt-get -qq update && \
    apt-get -qq install -y  

WORKDIR /app

COPY --from=build /websocket /websocket

EXPOSE 8080
CMD ["/websocket"]
