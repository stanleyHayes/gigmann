# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.25-alpine AS build
WORKDIR /src
# Cache modules first
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Static, stripped binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

# ---- runtime stage (distroless, non-root) ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
COPY --from=build /out/worker /worker
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
