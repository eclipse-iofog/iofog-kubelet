FROM golang:1.12-alpine as builder


ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

COPY . /go/src/github.com/eclipse-iofog/iofog-kubelet
WORKDIR /go/src/github.com/eclipse-iofog/iofog-kubelet
ARG BUILD_TAGS="netgo osusergo"

RUN apk add --update --no-cache bash curl git make && \
    make vendor && \
    make VK_BUILD_TAGS="${BUILD_TAGS}" build && \
    cp bin/iofog-kubelet /usr/bin/iofog-kubelet

FROM scratch
COPY --from=builder /usr/bin/iofog-kubelet /usr/bin/iofog-kubelet
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "/usr/bin/iofog-kubelet" ]
CMD [ "--help" ]
