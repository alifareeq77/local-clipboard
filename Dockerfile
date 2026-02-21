# Stage 1: build Vue UI
FROM node:20-alpine AS vue
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci 2>/dev/null || npm install
COPY web/ .
RUN npm run build

# Stage 2: build Go binary
FROM golang:1.22-alpine AS gobuilder
WORKDIR /app
COPY . .
RUN go mod download && CGO_ENABLED=0 go build -o /local-clipboard .

# Stage 3: minimal runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=gobuilder /local-clipboard .
COPY --from=vue /app/web/dist ./static
EXPOSE 8080
VOLUME /data
ENTRYPOINT ["/app/local-clipboard", "server", "-addr", ":8080", "-db", "/data/clipboard.db", "-static", "/app/static", "-no-build"]
