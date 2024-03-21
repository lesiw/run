FROM scratch
ARG ARCH
COPY build/gx-linux-${ARCH} /gx
ENTRYPOINT ["/gx"]
