## Original from: https://github.com/jamesob/docker-bitcoind
FROM ubuntu:20.04
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

ENV FILENAME dashcore-0.15.0.0-x86_64-linux-gnu.tar.gz
ENV DOWNLOAD_URL https://github.com/dashpay/dash/releases/download/v0.15.0.0/$FILENAME

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN mkdir /sources

WORKDIR /sources

RUN apt-get update

RUN apt-get install -y git

RUN git clone https://github.com/lomtax/digitalcoin.git

RUN cd /sources/digitalcoin

RUN apt-get update 

RUN DEBIAN_FRONTEND="noninteractive" apt-get install -y software-properties-common

RUN add-apt-repository -y ppa:pivx/pivx

RUN apt-get update

RUN apt-get install -y build-essential libtool autotools-dev autoconf pkg-config libssl-dev automake

RUN apt-get install -y libboost-all-dev libdb5.3-dev libdb5.3++-dev libprotobuf-dev protobuf-compiler libevent-dev libminiupnpc-dev libqrencode-dev

RUN DEBIAN_FRONTEND="noninteractive" DEBIAN_FRONTEND="noninteractive" apt-get install -y libqt5gui5 libqt5core5a libqt5dbus5 qttools5-dev qttools5-dev-tools

WORKDIR /sources/digitalcoin

RUN ./autogen.sh

RUN ./configure --enable-upnp-default --with-miniupnpc    --with-incompatible-bdb

RUN make install
EXPOSE 7998

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
