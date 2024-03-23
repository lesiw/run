FROM scratch
ARG ARCH
COPY build/pb-linux-${ARCH} /pb
ENTRYPOINT ["/pb"]
