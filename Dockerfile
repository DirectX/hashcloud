FROM golang:alpine

WORKDIR /go/src/github.com/DirectX/hashcloud
COPY . .

RUN apk add --update --no-cache git
RUN go get github.com/pilu/fresh
RUN go get
CMD ["fresh"]
