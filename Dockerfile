FROM golang
MAINTAINER Anders Pearson <anders@columbia.edu>
RUN apt-get update && apt-get install -y ca-certificates
RUN go get github.com/garyburd/go-oauth/oauth
RUN go get github.com/gorilla/feeds
RUN go get github.com/stvp/go-toml-config
RUN go get github.com/xiam/twitter
ADD . /go/src/github.com/thraxil/intweet
RUN go install github.com/thraxil/intweet
RUN mkdir /intweet/
EXPOSE 8890
CMD ["/go/bin/intweet", "-config=/intweet/config.toml"]

