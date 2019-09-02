## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

ENV FILENAME zen-2.0.21-1-amd64.deb
ENV DOWNLOAD_URL https://github.com/HorizenOfficial/zen/releases/download/v2.0.21-1/${FILENAME}

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install wget libc6 libgcc1 libgomp1 libstdc++6 && mkdir /download \
  && cd /download \
  && wget $DOWNLOAD_URL \
  && dpkg -i $FILENAME 

EXPOSE 8231

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
