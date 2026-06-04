FROM golang:1.23-alpine AS build

ENV GOTOOLCHAIN=auto

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/genzite-backend ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/genzite-backend /app/genzite-backend
COPY migrations /app/migrations

ENV PORT=8080
EXPOSE 8080

CMD ["/app/genzite-backend"]
