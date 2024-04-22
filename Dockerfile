FROM scratch
ARG ARCH
COPY out/run-linux-${ARCH} /run
ENTRYPOINT ["/run"]
