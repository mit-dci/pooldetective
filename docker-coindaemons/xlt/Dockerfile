## Original from: https://github.com/jamesob/docker-bitcoind
FROM ubuntu:20.04 as build-env
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y install libdb5.3 libdb5.3++  libdb5.3-dev libdb5.3++-dev libssl-dev libevent-dev  libboost-all-dev git apt-utils build-essential libtool autotools-dev automake pkg-config bsdmainutils curl software-properties-common cmake libcap-dev librsvg2-bin libz-dev libbz2-dev libtiff-tools python3-pip imagemagick
RUN mkdir /sources && cd /sources && git clone https://github.com/nexalt-project/nexalt-core
WORKDIR /sources/nexalt-core
RUN mkdir -p /sources/nexalt-core/out
RUN ./autogen.sh
WORKDIR /sources/nexalt-core
RUN ./configure --disable-tests --with-incompatible-bdb --disable-bench --bindir=/sources/nexalt-core/out/bin --libdir=/sources/nexalt-core/out/lib
RUN make -j8 install

FROM ubuntu:20.04
RUN apt-get update && apt-get -y install libdb5.3 libdb5.3++ libssl1.1 libevent-2.1-7 libevent-pthreads-2.1-7 libboost-system1.71 libboost-filesystem1.71 libboost-chrono1.71 libboost-thread1.71
COPY --from=build-env /sources/nexalt-core/out/bin/nexalt* /usr/local/bin/
RUN mkdir /root/.nexalt
EXPOSE 9332

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
