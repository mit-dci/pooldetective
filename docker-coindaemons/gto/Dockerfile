## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

ENV FILENAME globaltoken-3.2.1-x86_64-linux-gnu.tar.gz
ENV DOWNLOAD_URL https://globaltoken.org/download.php?os=x86_64-linux-gnu

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install wget && mkdir /download \
  && cd /download \
  && wget $DOWNLOAD_URL -O $FILENAME \
  && tar -xvf $FILENAME \
  && rm -rf $FILENAME \
  && mkdir /root/.globaltoken \
  && mv /download/bin/* /usr/local/bin/ 

EXPOSE 9320 19320

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]

