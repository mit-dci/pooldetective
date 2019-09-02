## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

ENV FILENAME raven-3.3.2.0-x86_64-linux-gnu.tar.gz
ENV DOWNLOAD_URL https://github.com/RavenProject/Ravencoin/releases/download/v3.3.2/${FILENAME}

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install wget && mkdir /download \
  && cd /download \
  && wget $DOWNLOAD_URL \
  && tar -xvf $FILENAME \
  && rm -rf $FILENAME \
  && mkdir /root/.raven \
  && mv /download/raven-3.3.2.0/bin/* /usr/local/bin/ 

EXPOSE 8766

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
