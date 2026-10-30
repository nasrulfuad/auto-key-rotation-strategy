FROM docker.io/library/golang:1.26-alpine AS dev
RUN go install github.com/air-verse/air@v1.62.0
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
CMD ["air", "-c", ".air.toml"]

FROM docker.io/library/golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM docker.io/library/alpine:3.22
RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/server /usr/local/bin/server
ENTRYPOINT ["/usr/local/bin/server"]
