FROM golang:1.24.1 as BUILDER
RUN go env -w GOPROXY=https://goproxy.cn,direct

ARG GH_USER
ARG GH_TOKEN
RUN echo "machine github.com login ${GH_USER} password ${GH_TOKEN}" > $HOME/.netrc

# build binary
WORKDIR /go/src/github.com/opensourceways/image-scanning
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o image-scanning .

# copy binary config and utils
FROM openeuler/openeuler:24.03-lts-sp2
RUN dnf -y update --repo OS --repo update && \
    dnf in -y shadow --repo OS --repo update && \
    dnf install -y git golang docker-ce && \
    dnf remove -y gdb-gdbserver && \
    groupadd -g 1000 image-scanning  && \
    useradd -u 1000 -g image-scanning -s /sbin/nologin -m image-scanning && \
    echo > /etc/issue && echo > /etc/issue.net && echo > /etc/motd && \
    echo "umask 027" >> /root/.bashrc &&\
    echo 'set +o history' >> /root/.bashrc && \
    sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs && \
    rm -rf /tmp/*

USER image-scanning

COPY  --chown=image-scanning --from=BUILDER /go/src/github.com/opensourceways/image-scanning/image-scanning /opt/app/image-scanning
COPY --chown=image-scanning --from=BUILDER /go/src/github.com/opensourceways/image-scanning/script/trivy_env.sh /opt/app/trivy_env.sh
COPY --chown=image-scanning --from=BUILDER /go/src/github.com/opensourceways/image-scanning/script/save_image.sh /opt/app/save_image.sh

RUN chmod 550 /opt/app/trivy_env.sh /opt/app/save_image.sh && mkdir images

WORKDIR /opt/app/

ENTRYPOINT ["/opt/app/image-scanning"]