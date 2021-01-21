## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim AS build-env
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get -y install git build-essential pkg-config libc6-dev m4 g++-multilib autoconf libtool ncurses-dev unzip git python python-zmq zlib1g-dev wget curl bsdmainutils automake
RUN mkdir /sources \
  && cd /sources \
  && git clone https://github.com/PirateNetwork/PirateOcean.git \
  && cd PirateOcean 
WORKDIR /sources/PirateOcean

FROM debian:stable-slim
COPY --from=build-env /sources/PirateOcean/zcutil/fetch-params.sh /usr/local/bin

RUN apt-get update && apt-get install -y unzip wget && wget https://github.com/PirateNetwork/SevenSeas/releases/download/0.8.2/sevenseas-linux-x86-0.8.2.zip \
  && unzip sevenseas-linux-x86-0.8.2.zip \
  && mv /SevenSeas/* /usr/local/bin/ 

EXPOSE 8332 18332

ADD ./pirated /usr/local/bin/pirated
RUN chmod a+x /usr/local/bin/pirated
ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/fetch-params.sh
RUN apt-get install -y libgomp1
ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
