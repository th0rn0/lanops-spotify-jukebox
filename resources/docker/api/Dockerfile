FROM golang:1.22-alpine
LABEL maintainer="Thornton Phillis (dev@th0rn0.co.uk)"

WORKDIR /app

ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN apk add --no-cache gcc musl-dev

COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ .

RUN go build -ldflags='-s -w -extldflags "-static"' -o /spotify-api

EXPOSE 8080

CMD [ "/spotify-api" ]