## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"
ARG DEBIAN_FRONTEND=noninteractive
ENV FILENAME pirl-linux-amd64-1.8.27-damocles.tar.gz
ENV DOWNLOAD_URL https://git.pirl.io/community/pirl/uploads/b05e46f46d2a5e843d02cac13ef64c3d/${FILENAME}

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install wget unzip && mkdir /download \
  && cd /download \
  && wget $DOWNLOAD_URL \
  && tar -xvf $FILENAME \
  && rm -rf $FILENAME \
  && mkdir /root/.pirl \
  && mv /download/pirl-linux-amd64-1.8.27-damocles/* /usr/local/bin/ 

EXPOSE 8585

ENTRYPOINT ["pirl-linux-amd64-1.8.27-damocles","--rpcaddr","0.0.0.0"]
