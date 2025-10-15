FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./
RUN apk update && apk add --no-cache git

ARG GH_PAT

# Configure git for private repositories
RUN git config --global url."https://${GH_PAT}:@github.com/".insteadOf "https://github.com/"

# Set GOPRIVATE to ensure Go doesn't try to use proxy for private repos
ENV GOPRIVATE=github.com/hivemindd/*
ENV GOPROXY=direct

# Download dependencies only (cached if go.mod/go.sum unchanged)
RUN go mod download

# Now copy the rest of the source code
COPY . .

RUN go build -o app .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/templates ./templates
EXPOSE 9002

CMD ["./app"] 
