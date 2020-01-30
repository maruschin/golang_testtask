FROM debian:stretch

RUN apt-get update && apt-get install -y \
	golang-go \
	redis-server \
	git

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

RUN go get -u github.com/go-redis/redis
RUN go get -u github.com/gorilla/mux


RUN redis-server && go run main.go
