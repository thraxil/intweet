FROM ubuntu:trusty
MAINTAINER Anders Pearson <anders@columbia.edu>
RUN apt-get update
RUN apt-get install -y ca-certificates
RUN mkdir /intweet/
ADD ./intweet /intweet/
CMD ["/intweet/intweet", "-config=/intweet/config.toml"]
EXPOSE 8890
