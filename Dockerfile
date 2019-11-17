FROM golang:1.12-alpine as builder

RUN apk add --update --no-cache bash curl git make

COPY *go* /
COPY version /
COPY AUTHORS /
COPY Makefile /
COPY netlify.toml /
COPY ./cmd /cmd
COPY ./log /log
COPY ./hack /hack
COPY ./.git /.git
COPY ./trace /trace
COPY ./manager /manager
COPY ./versions /versions
COPY ./vkubelet /vkubelet
COPY ./providers /providers

ARG BUILD_TAGS="netgo osusergo"

WORKDIR /

RUN . version && export MAJOR && export MINOR && export PATCH && export SUFFIX && make VK_BUILD_TAGS="${BUILD_TAGS}" build

FROM scratch
COPY --from=builder /bin/iofog-kubelet /usr/bin/iofog-kubelet
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "/usr/bin/iofog-kubelet" ]
CMD [ "--help" ]
