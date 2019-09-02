## Original from: https://github.com/jamesob/docker-bitcoind
FROM ubuntu:18.04
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

ENV FILENAME ImageCoin_linux_32_64.tar.xz
ENV DOWNLOAD_URL https://github.com/mceme/ImageCoin/releases/download/1.0.33/${FILENAME}

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install wget xz-utils
RUN mkdir /download \
  && cd /download \
  && wget $DOWNLOAD_URL \
  && tar -xvf $FILENAME \
  && rm -rf $FILENAME \
  && mkdir /root/.imagecoin \
  && mv /download/* /usr/local/bin/ 

RUN apt-get -y install libzmq5 libboost-system-dev libboost-filesystem-dev
RUN apt-get -y install libboost-program-options-dev
RUN apt-get -y install libboost-thread-dev
RUN apt-get -y install software-properties-common
RUN add-apt-repository ppa:bitcoin/bitcoin
RUN apt-get -y install libdb4.8-dev libdb4.8++-dev
RUN apt-get -y install libminiupnpc-dev libcrypto++-dev libssl-dev 
RUN apt-get -y install libevent-dev
EXPOSE 9337

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
