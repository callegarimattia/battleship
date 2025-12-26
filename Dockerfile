# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies if needed (e.g. git, make)
# RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary
# -ldflags="-w -s" reduces binary size by stripping debug info
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o battleship ./cmd/server

# Final Stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /app/battleship /battleship

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/battleship"]
