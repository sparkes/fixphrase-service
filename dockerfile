# syntax=docker/dockerfile:1

# -------- Build stage --------
FROM golang:1.22-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/fixphrase ./main.go

# -------- Runtime stage --------
FROM gcr.io/distroless/static:nonroot
WORKDIR /

# Allow configurable port at build-time (default 7080)
ARG PORT=7080
ENV PORT=${PORT}

# The service itself reads ADDR from environment (.env or runtime env vars)
# If ADDR is not set, the Go service defaults to :7080 internally

COPY --from=build /out/fixphrase /fixphrase

# Expose configured port (metadata only)
EXPOSE ${PORT}

USER nonroot:nonroot
ENTRYPOINT ["/fixphrase"]