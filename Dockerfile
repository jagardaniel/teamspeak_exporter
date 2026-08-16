# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s" -o /teamspeak_exporter

FROM gcr.io/distroless/static-debian13:nonroot AS release

WORKDIR /app

COPY --from=build /teamspeak_exporter /app/teamspeak_exporter

EXPOSE 9800

ENTRYPOINT ["/app/teamspeak_exporter"]
