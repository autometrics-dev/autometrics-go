FROM golang:1.20-alpine
MAINTAINER Fiberplane <info@fiberplane.com>
ARG version=development

RUN apk update && apk add git

WORKDIR /app

COPY . ./

RUN go mod download

WORKDIR /app/examples/web

RUN go generate cmd/main.go

ENV VERSION="$version"

RUN scripts/build_server

RUN cp web-server /

EXPOSE 62086

CMD [ "/web-server" ]
