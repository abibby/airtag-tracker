# syntax=docker/dockerfile:1

FROM golang:1.22 as build

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY . ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /airtag-tracker

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 8080

FROM debian:bookworm

RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y ssh tesseract-ocr ca-certificates \
    && apt-get clean

COPY --from=build /airtag-tracker /airtag-tracker

# Run
CMD ["/airtag-tracker"]