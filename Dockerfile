FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/messenger ./cmd/messenger

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/messenger /usr/local/bin/messenger
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/messenger"]