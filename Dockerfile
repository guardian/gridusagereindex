FROM golang
ADD . /go/src/github.com/guardian/gridusagereindex
ADD config.json /etc/gridusagereindex.json
RUN go install github.com/guardian/gridusagereindex
ENTRYPOINT /go/bin/gridusagereindex
EXPOSE 8080
