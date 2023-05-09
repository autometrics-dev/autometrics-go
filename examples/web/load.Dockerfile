FROM golang:1.20-alpine
MAINTAINER Fiberplane <info@fiberplane.com>

RUN apk update && apk add curl

COPY examples/web/scripts/poll_server /

CMD [ "/poll_server" ]
