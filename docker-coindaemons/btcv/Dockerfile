## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable as build-env
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get -y install git apt-utils build-essential libtool autotools-dev automake pkg-config bsdmainutils curl software-properties-common cmake libcap-dev librsvg2-bin libz-dev libbz2-dev libtiff-tools python3-pip imagemagick
RUN mkdir /sources && cd /sources && git clone https://github.com/bitcoinvault/bitcoinvault
WORKDIR /sources/bitcoinvault
RUN make -j8 -C depends HOST=x86_64-unknown-linux-gnu
RUN mkdir -p /sources/bitcoinvault/out
RUN ./autogen.sh
RUN mkdir build
WORKDIR /sources/bitcoinvault/build
RUN ../configure --disable-dependency-tracking --disable-tests --disable-bench --enable-glibc-back-compat --prefix=/sources/bitcoinvault/depends/x86_64-unknown-linux-gnu --enable-reduce-exports --bindir=/sources/bitcoinvault/out/bin --libdir=/sources/bitcoinvault/out/lib
RUN make -j8 install

FROM debian:stable-slim
COPY --from=build-env /sources/bitcoinvault/out/bin/bvault* /usr/local/bin/
RUN mkdir /root/.bvault
EXPOSE 3332

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
