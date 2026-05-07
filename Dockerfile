# Multi-stage build for PikaAnalytics
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Build backend
FROM golang:1.24-alpine AS backend-builder

ARG VERSION=unknown

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
COPY --from=frontend-builder /app/frontend/build ./frontend

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o pikaanalytics main.go

RUN echo -n "$VERSION" > version.txt

# Final stage
FROM alpine:latest

ARG VERSION=unknown

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

COPY --from=backend-builder /app/backend/pikaanalytics .
COPY --from=backend-builder /app/backend/frontend ./frontend
COPY --from=backend-builder /app/backend/version.txt .

RUN mkdir -p /app/data

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/admin/ || exit 1

CMD ["./pikaanalytics"]
