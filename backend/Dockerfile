# ── Build stage ─────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS build
WORKDIR /app

# Cache deps first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static binary so it runs on a distroless image.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

# ── Run stage ───────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /server /server
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
