FROM golang:1.5.2

ADD . /go/src/github.com/walesey/goprox
WORKDIR /go/src/github.com/walesey/goprox

EXPOSE 80 
ENV PORT 80 

RUN go build

CMD ./goprox
