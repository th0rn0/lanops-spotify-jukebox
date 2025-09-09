FROM golang:1.24-alpine
LABEL maintainer="Thornton Phillis (dev@th0rn0.co.uk)"

WORKDIR /app

ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN apk add --no-cache gcc musl-dev

COPY service/go.mod service/go.sum ./
RUN go mod download

COPY service/ .

RUN go build -ldflags='-s -w -extldflags "-static"' -o bin/spotify-jukebox ./cmd/spotify-jukebox

EXPOSE 8888

CMD [ "bin/spotify-jukebox" ]