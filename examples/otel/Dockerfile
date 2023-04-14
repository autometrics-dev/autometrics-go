FROM golang:1.20-alpine
MAINTAINER Fiberplane <info@fiberplane.com>

# Cannot really build the demo image from
# the examples subfolder because of
# relative imports shenanigans that go out of build context (i.e. upwards)
#
# Use
# GOOS=linux GOARCH=amd64 go build -o web-server ./cmd/main.go
#
# To build the web-server app

COPY web-server /

EXPOSE 62086

CMD [ "/web-server" ]
