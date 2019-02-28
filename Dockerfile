FROM golang:alpine

WORKDIR /app
COPY . .

RUN apk add --update --no-cache git gcc libc-dev
RUN go get
RUN go build

EXPOSE 3010

CMD ["./hashcloud"]
