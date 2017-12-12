FROM ubuntu:16.04

MAINTAINER karlmutch@gmail.com

LABEL vendor="Sentient Technologies INC" \
      ai.sentient.version=0.0.0 \
      ai.sentient.module=studio-go-runner

ENV LANG C.UTF-8

ARG USER
ENV USER ${USER}
ARG USER_ID
ENV USER_ID ${USER_ID}
ARG USER_GROUP_ID
ENV USER_GROUP_ID ${USER_GROUP_ID}

ENV GO_VERSION 1.9.2

RUN apt-get -y update


RUN \
    apt-get -y install software-properties-common wget openssl ssh curl jq apt-utils && \
    apt-get -y install make git gcc && \
    apt-get clean && \
    apt-get autoremove && \
    groupadd -f -g ${USER_GROUP_ID} ${USER} && \
    useradd -g ${USER_GROUP_ID} -u ${USER_ID} -ms /bin/bash ${USER}

USER ${USER}
WORKDIR /home/${USER}

RUN cd /home/${USER} && \
    mkdir -p /home/${USER}/go && \
    wget -O /tmp/go.tgz https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar xzf /tmp/go.tgz && \
    rm /tmp/go.tgz


ENV PATH=$PATH:/home/${USER}/go/bin
ENV GOROOT=/home/${USER}/go
ENV GOPATH=/project

VOLUME /project
WORKDIR /project/src/github.com/KarlMutch/MeshTest

CMD /bin/bash -C ./build.sh
