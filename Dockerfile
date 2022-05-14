FROM golang:1.18-alpine As build
WORKDIR /app
ENV CGO_ENABLED=0
COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN go mod download
COPY . /app/
RUN go build -buildvcs=false -o=/out/notion-front

FROM alpine:latest
ENV LISTEN_ADDR=0.0.0.0:80
ENV SOURCE_DIR=/content
ENV CACHE_DIR=/cache
COPY docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=build /out/notion-front /opt/notion-front/notion-front
COPY --from=build /app/www /opt/notion-front/www
ENTRYPOINT [ "/docker-entrypoint.sh" ]
