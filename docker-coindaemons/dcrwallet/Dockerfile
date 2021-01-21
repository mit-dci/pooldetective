#DOCKER_IMAGE_TAG=dcrd-mainnet
#mkdir .dcrd
#docker run -d --network=host -P -v $(pwd)/.dcrd/:/home/decred/.dcrd/ decred/$DOCKER_IMAGE_TAG

FROM ubuntu:xenial

LABEL description="Decred mainnet daemon"
LABEL version="1.3"
LABEL maintainer="john@netpurgatory.com"

ENV DCR_RELEASE v1.4.0
ENV TERM linux
ENV USER decred
ENV HOME /home/$USER
ENV DOTDCRD $HOME/.dcrd

# create user
RUN adduser --disabled-password --gecos '' $USER
ARG DEBIAN_FRONTEND=noninteractive
RUN set -x && \

    # Update base distro & install build tooling
    DEBIAN_FRONTEND=noninteractive && \
    BUILD_DEPS=curl && \
    apt-get update && \
    apt-get install -qy $BUILD_DEPS && \

    # Decred release url and files
    DCR_RELEASE_URL="https://github.com/decred/decred-binaries/releases/download/$DCR_RELEASE" && \
    DCR_MANIFEST_FILE="manifest-$DCR_RELEASE.txt" && \
    DCR_RELEASE_NAME="decred-linux-amd64-$DCR_RELEASE" && \
    DCR_RELEASE_FILE="$DCR_RELEASE_NAME.tar.gz" && \

    # Import Decred gpg key
    gpg --keyserver pgp.mit.edu --recv-keys 0x518A031D && \

    # Download archives
    cd /tmp && \
    curl -SLO $DCR_RELEASE_URL/$DCR_RELEASE_FILE && \
    curl -SLO $DCR_RELEASE_URL/$DCR_MANIFEST_FILE && \
    curl -SLO $DCR_RELEASE_URL/$DCR_MANIFEST_FILE.asc && \

    # Verify signature and hash
    gpg --verify --trust-model=always $DCR_MANIFEST_FILE.asc && \
    grep "$DCR_RELEASE_FILE" $DCR_MANIFEST_FILE | sha256sum -c - && \
    rm -R $HOME/.gnupg && \

    # Extract and install
    tar xvzf $DCR_RELEASE_FILE && \
    mv $DCR_RELEASE_NAME/dcrd $HOME && \
    mkdir -p $DOTDCRD && \
    chown -R $USER.$USER $HOME && \

    # Cleanup
    apt-get -qy remove $BUILD_DEPS && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

# switch user for runtime
USER $USER
WORKDIR /home/$USER

# PEER & RPC PORTS
EXPOSE 9108 9109

CMD "/usr/local/bin/docker_entrypoint.sh"
