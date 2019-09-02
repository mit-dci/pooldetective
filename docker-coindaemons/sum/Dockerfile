## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable as build-env
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

RUN apt-get update && apt-get -y install git apt-utils build-essential libtool autotools-dev automake pkg-config bsdmainutils curl software-properties-common cmake libcap-dev librsvg2-bin libz-dev libbz2-dev libtiff-tools python3-pip imagemagick
RUN mkdir /sources && cd /sources && git clone https://github.com/sumcoinlabs/sumcoin
WORKDIR /sources/sumcoin
RUN make -j8 -C depends HOST=x86_64-unknown-linux-gnu
RUN mkdir -p /sources/sumcoin/out
RUN ./autogen.sh
RUN mkdir build
WORKDIR /sources/sumcoin/build
RUN ../configure --disable-dependency-tracking --disable-tests --disable-bench --enable-glibc-back-compat --prefix=/sources/sumcoin/depends/x86_64-unknown-linux-gnu --enable-reduce-exports --bindir=/sources/sumcoin/out/bin --libdir=/sources/sumcoin/out/lib
RUN make -j8 install

FROM debian:stable-slim
COPY --from=build-env /sources/sumcoin/out/bin/sum* /usr/local/bin/
RUN mkdir /root/.sumcoin
EXPOSE 3332

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
