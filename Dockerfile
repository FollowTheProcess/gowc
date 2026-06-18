FROM gcr.io/distroless/static
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/gowc /usr/local/bin/gowc
ENTRYPOINT [ "/usr/local/bin/gowc" ]
