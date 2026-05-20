FROM golang:1.26.1-alpine AS build

ARG GIT_REPO_URL
ARG GIT_BRANCH

WORKDIR /src
RUN apk add --no-cache ca-certificates git

RUN git clone --depth 1 --branch "${GIT_BRANCH}" "${GIT_REPO_URL}" /src/repo
WORKDIR /src/repo/oauth
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/oauth .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/oauth /app/oauth

ENV DATA_DIR=/data
EXPOSE 80
VOLUME ["/data"]
ENTRYPOINT ["/app/oauth"]
