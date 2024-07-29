# syntax=docker/dockerfile:1

FROM golang:1.23-rc AS builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/scraper


FROM node:21-alpine

RUN npm install -g json-server

WORKDIR /app

RUN mkdir db/

COPY --from=builder /app/scraper /app/scraper
COPY run.sh /app

RUN chmod +x /app/run.sh

EXPOSE 3000

# Run
CMD ["/bin/sh", "-c", "/app/run.sh"]
