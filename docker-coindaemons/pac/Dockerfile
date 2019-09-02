## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stretch
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

ENV FILENAME pacglobal-v0.14.0.6-d6fbc8bb24-legacylin64.tgz
ENV DOWNLOAD_URL https://github.com/PACGlobalOfficial/PAC/releases/download/d6fbc8bb24/${FILENAME}

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y upgrade && apt-get -y install wget && mkdir /download \
  && cd /download \
  && wget $DOWNLOAD_URL \
  && tar -xvf $FILENAME \
  && rm -rf $FILENAME \
  && mkdir /root/.pac \
  && mv /download/pacglobal-v0.14.0.6-d6fbc8bb24-legacylin64/* /usr/local/bin/ 

EXPOSE 8212

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
