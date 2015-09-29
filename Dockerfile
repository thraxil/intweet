FROM centurylink/ca-certs
MAINTAINER Anders Pearson <anders@columbia.edu>
COPY intweet /
EXPOSE 8890
CMD ["/intweet", "-config=/config.toml"]

