# Docker multi stage build formatted file.  This is used to build then prepare
# containers for the services that this repository uses
#
FROM golang:1.11.10

MAINTAINER karlmutch@gmail.com

LABEL vendor="Cognizant Technologies" \
      dev.cognizant-ai.version=0.0.0 \
      dev.cognizant-ai.module=platform-services

ENV LANG C.UTF-8

# Protobuf version
ENV PROTOBUF_VERSION="3.7.1"
ENV PROTOBUF_ZIP=protoc-${PROTOBUF_VERSION}-linux-x86_64.zip
ENV PROTOBUF_URL=https://github.com/google/protobuf/releases/download/v${PROTOBUF_VERSION}/${PROTOBUF_ZIP}

ARG USER
ENV USER ${USER}
ARG USER_ID
ENV USER_ID ${USER_ID}
ARG USER_GROUP_ID
ENV USER_GROUP_ID ${USER_GROUP_ID}

RUN apt-get -y update

RUN apt-get -y install git software-properties-common wget openssl ssh curl jq apt-utils unzip python-pip && \
    apt-get clean && \
    apt-get autoremove && \
    pip install awscli --upgrade

RUN groupadd -f -g ${USER_GROUP_ID} ${USER}
RUN useradd -g ${USER_GROUP_ID} -u ${USER_ID} -ms /bin/bash ${USER}

RUN wget ${PROTOBUF_URL} && \
    unzip ${PROTOBUF_ZIP} -d /usr && \
    chmod +x /usr/bin/protoc && \
    find /usr/include/google -type d -print0 | xargs -0 chmod ugo+rx && \
    chmod -R +r /usr/include/google

USER ${USER}
WORKDIR /home/${USER}

ENV GOPATH=/project
VOLUME /project
WORKDIR /project/src/github.com/leaf-ai/platform-services

CMD /bin/bash -C ./all-build.sh
