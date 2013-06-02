all: intweet

intweet: intweet.go
	go build intweet.go

install_deps:
	go get github.com/garyburd/go-oauth/oauth
	go get github.com/gorilla/feeds
	go get github.com/stvp/go-toml-config
	go get github.com/xiam/twitter
