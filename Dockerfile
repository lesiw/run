FROM scratch
ARG ARCH
COPY out/pb-linux-${ARCH} /pb
ENTRYPOINT ["/pb"]
