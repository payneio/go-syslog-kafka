FROM flynn/busybox
MAINTAINER Paul Payne <paul@payne.io>

ADD stage/go-syslog-kafka /bin/

# syslog
EXPOSE 514
EXPOSE 514/udp

ENTRYPOINT ["/bin/go-syslog-kafka"]
CMD []
