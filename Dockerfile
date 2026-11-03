FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /monitor ./cmd/monitor
FROM alpine:3.20
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /monitor /app/monitor
COPY configs /app/configs
USER app
EXPOSE 8080
ENTRYPOINT ["/app/monitor"]
