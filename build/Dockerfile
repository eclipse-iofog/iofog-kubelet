FROM golang:alpine as builder

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN apk add --no-cache \
	ca-certificates \
	--virtual .build-deps \
	git \
	gcc \
	libc-dev \
	libgcc \
	make \
	bash

COPY . /go/src/github.com/eclipse-iofog/iofog-kubelet
WORKDIR /go/src/github.com/eclipse-iofog/iofog-kubelet
ARG BUILD_TAGS="netgo osusergo"
RUN make VK_BUILD_TAGS="${BUILD_TAGS}" build
RUN cp bin/iofog-kubelet /usr/bin/iofog-kubelet


FROM scratch
COPY --from=builder /usr/bin/iofog-kubelet /usr/bin/iofog-kubelet
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs
ENTRYPOINT [ "/usr/bin/iofog-kubelet" ]
CMD [ "--help" ]
