FROM golang
ADD . /go/src/github.com/cduchesne/scaleio-graphite
RUN go get github.com/marpaia/graphite-golang
RUN go install github.com/cduchesne/scaleio-graphite
ENTRYPOINT /go/bin/scaleio-graphite
