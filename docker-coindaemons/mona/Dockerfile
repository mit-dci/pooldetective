## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"
ARG DEBIAN_FRONTEND=noninteractive
ENV FILENAME monacoin-0.17.1-x86_64-linux-gnu.tar.gz
ENV DOWNLOAD_URL https://github.com/monacoinproject/monacoin/releases/download/monacoin-0.17.1/${FILENAME}

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install wget unzip && mkdir /download \
  && cd /download \
  && wget $DOWNLOAD_URL \
  && tar -xvf $FILENAME \
  && rm -rf $FILENAME \
  && mkdir /root/.monacoin \
  && mv /download/monacoin-0.17.1/bin/* /usr/local/bin/ 

EXPOSE 9402

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
