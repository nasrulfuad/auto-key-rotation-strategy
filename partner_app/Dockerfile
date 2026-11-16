FROM docker.io/library/golang:1.27.3-alpine3.24@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS dev
RUN go install github.com/air-verse/air@v1.62.0
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
CMD ["air", "-c", ".air.toml"]

FROM docker.io/library/golang:1.27.3-alpine3.24@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM docker.io/library/alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf AS runtime
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app
USER 10001
COPY --from=build /out/server /usr/local/bin/server
ENTRYPOINT ["/usr/local/bin/server"]
