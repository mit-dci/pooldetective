## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"
ARG DEBIAN_FRONTEND=noninteractive
ENV FILENAME atom-0.16.2-x86_64-linux-gnu.tar.gz
ENV DOWNLOAD_URL https://github.com/bitcoin-atom/bitcoin-atom/releases/download/v0.16.2/${FILENAME}

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y install wget unzip && mkdir /download \
  && cd /download \
  && wget $DOWNLOAD_URL \
  && tar -xvf $FILENAME \
  && rm -rf $FILENAME \
  && mkdir /root/.dogecoin \
  && mv /download/atom-0.16.2/bin/* /usr/local/bin/ 

EXPOSE 7332

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
