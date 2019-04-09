ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: intweet

intweet: intweet.go
	go build intweet.go

install_deps:
	go get github.com/garyburd/go-oauth/oauth
	go get github.com/gorilla/feeds
	go get github.com/stvp/go-toml-config
	go get github.com/ChimeraCoder/anaconda

build: intweet
	docker run --rm -v $(ROOT_DIR):/src -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder thraxil/intweet

push:
	docker push thraxil/intweet
