# Docker multi stage build formatted file.  This is used to build then prepare
# containers for the services that this repository uses
#
FROM golang:1.9.2

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

RUN apt-get -y update

RUN apt-get -y install git software-properties-common wget openssl ssh curl jq apt-utils && \
    apt-get clean && \
    apt-get autoremove && \
    groupadd -f -g ${USER_GROUP_ID} ${USER} && \
    useradd -g ${USER_GROUP_ID} -u ${USER_ID} -ms /bin/bash ${USER}

USER ${USER}
WORKDIR /home/${USER}

ENV GOPATH=/project

VOLUME /project
WORKDIR /project/src/github.com/karlmutch/MeshTest

CMD /bin/bash -C ./cmd/timesrv/build.sh ; /bin/bash -C ./cmd/expmanager/build.sh 
