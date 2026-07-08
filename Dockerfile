FROM golang:1.23-alpine AS build
WORKDIR /src

COPY oauth/go.mod oauth/go.sum ./
RUN go mod download

COPY oauth/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/oauth .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/oauth /app/oauth

ENV DATA_DIR=/data
EXPOSE 80
VOLUME ["/data"]
ENTRYPOINT ["/app/oauth"]
