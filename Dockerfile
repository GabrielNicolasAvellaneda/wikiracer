FROM golang:1.8

MAINTAINER maksym.naboka@gmail.com

ADD . /go/src/github.com/darkonie/wikiracer

RUN cd /go/src/github.com/darkonie/wikiracer && go install

ENTRYPOINT /go/bin/wikiracer

EXPOSE 8081
