## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

ENV FILENAME vertcoind-v0.14.0-linux-amd64.zip
ENV DOWNLOAD_URL https://github.com/vertcoin-project/vertcoin-core/releases/download/0.14.0/${FILENAME}

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install wget unzip && mkdir /download \
  && cd /download \
  && wget $DOWNLOAD_URL \
  && unzip $FILENAME \
  && rm -rf $FILENAME \
  && mkdir /root/.vertcoin \
  && mv /download/* /usr/local/bin/ 

EXPOSE 5888

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
