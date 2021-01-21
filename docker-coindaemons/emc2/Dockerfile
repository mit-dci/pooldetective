## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"
ARG DEBIAN_FRONTEND=noninteractive
ENV FILENAME EinsteiniumLightsaber-linux.tar.gz
ENV DOWNLOAD_URL https://download.emc2.foundation/wallets/${FILENAME}

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install wget && mkdir /download \
  && cd /download \
  && wget --no-check-certificate $DOWNLOAD_URL \
  && tar -xvf $FILENAME \
  && rm -rf $FILENAME \
  && mkdir /root/.emc2 \
  && mv /download/* /usr/local/bin/ 

EXPOSE 8766

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
