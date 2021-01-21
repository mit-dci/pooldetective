## Original from: https://github.com/jamesob/docker-bitcoind
FROM alpine
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

ARG GLIBC_VERSION=2.29-r0
ARG DEBIAN_FRONTEND=noninteractive
ENV FILENAME bitcoin-cash-node-22.1.0-x86_64-linux-gnu.tar.gz
ENV DOWNLOAD_URL https://github.com/bitcoin-cash-node/bitcoin-cash-node/releases/download/v22.1.0/$FILENAME

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apk update \
  && apk --no-cache add wget tar bash ca-certificates \
  && wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub \
  && wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VERSION}/glibc-${GLIBC_VERSION}.apk \
  && wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VERSION}/glibc-bin-${GLIBC_VERSION}.apk \
  && apk --no-cache add glibc-${GLIBC_VERSION}.apk \
  && apk --no-cache add glibc-bin-${GLIBC_VERSION}.apk \
  && rm -rf /glibc-${GLIBC_VERSION}.apk \
  && rm -rf /glibc-bin-${GLIBC_VERSION}.apk \
  && wget $DOWNLOAD_URL \
  && tar xzvf $FILENAME \
  && mkdir /root/.bitcoin \
  && mv /bitcoin-cash-node-22.1.0/bin/* /usr/local/bin/ \
  && rm -rf /bitcoin-cash-node-22.1.0/ \
  && rm -rf $FILENAME \
  && apk del tar wget ca-certificates

EXPOSE 8332 18332

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
