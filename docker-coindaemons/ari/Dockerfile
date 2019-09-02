## Original from: https://github.com/jamesob/docker-bitcoind
FROM rgrubba/debian-squeeze as build-env
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y --force-yes install libdb4.8-dev libdb4.8++-dev libssl-dev libevent-dev  libboost-all-dev git apt-utils build-essential libtool autotools-dev automake pkg-config bsdmainutils curl cmake libcap-dev librsvg2-bin libz-dev libbz2-dev libtiff-tools imagemagick
RUN mkdir /sources && cd /sources && git clone https://github.com/AricoinCurrency/Aricoin
RUN apt-get -y install libminiupnpc-dev
WORKDIR /sources/Aricoin/src
RUN make -j8 -f makefile.unix USE_UPNP=-

FROM rgrubba/debian-squeeze
RUN apt-get update && apt-get -y install libdb4.8 libdb4.8++ libssl libevent-2.1-6 libevent-pthreads-2.1-6 libboost-program-options1.42 libboost-system1.42 libboost-filesystem1.42 libboost-chrono1.42 libboost-thread1.42
COPY --from=build-env /sources/Aricoin/src/aricoin* /usr/local/bin/
RUN mkdir /root/.aricoin
EXPOSE 3332

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]

