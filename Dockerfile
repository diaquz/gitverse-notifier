FROM golang:1.26.1-trixie AS stage-build

ARG TARGETARCH
ARG VERSION
ARG CHECK_VERSION=1.0.9

ENV VERSION=$VERSION
ENV GOPATH=/go
ENV PATH=/go/bin:/usr/local/go/bin:$PATH
ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /opt/gitverse_notifier
COPY pkg pkg
COPY go* .
COPY main.go .
COPY config.yml.example .
COPY entrypoint.sh .
COPY Makefile .

RUN go mod download -x

RUN make build -s \
    && set -x && ls -al . \
    && mv /opt/gitverse_notifier/build/gitverse_notifier /opt/gitverse_notifier/gitverse_notifier

RUN mkdir /opt/gitverse_notifier/release \
    && mv /opt/gitverse_notifier/config_example.yml /opt/gitverse_notifier/release \
    && mv /opt/gitverse_notifier/entrypoint.sh /opt/gitverse_notifier/release \
    && chmod 755 /opt/gitverse_notifier/release/entrypoint.sh 

FROM debian:trixie
ARG TARGETARCH
ENV LANG=en_US.UTF-8

# LABEL org.opencontainers.image.source
LABEL org.opencontainers.image.description="Gitverse notifier"

ARG DEPENDENCIES="                    \
        bash-completion               \
        jq                            \
        less                          \
        redis-tools                   \
        ca-certificates"

ARG APT_MIRROR=http://deb.debian.org

RUN set -ex \
    && sed -i "s@http://.*.debian.org@${APT_MIRROR}@g" /etc/apt/sources.list.d/debian.sources \
    && ln -sf /usr/share/zoneinfo/Europe/Moscow /etc/localtime \
    && apt-get update \
    && apt-get install -y --no-install-recommends ${DEPENDENCIES} \
    && apt-get clean all \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/gitverse_notifier

COPY --from=stage-build /opt/gitverse_notifier/release .
COPY --from=stage-build /opt/gitverse_notifier/gitverse_notifier .

ARG VERSION
ENV VERSION=${VERSION}

VOLUME /opt/gitverse_notifier/data

ENTRYPOINT ["./entrypoint.sh"]

STOPSIGNAL SIGQUIT

CMD [ "./gitverse_notifier" ]
