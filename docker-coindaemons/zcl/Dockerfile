## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim AS build-env
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install git build-essential pkg-config libc6-dev m4 g++-multilib autoconf libtool ncurses-dev unzip git python python-zmq zlib1g-dev wget curl bsdmainutils automake
RUN mkdir /sources \
  && cd /sources \
  && git clone https://github.com/ZclassicCommunity/zclassic.git \
  && cd zclassic \
  && git checkout v2.1.1-2
WORKDIR /sources/zclassic
RUN ./zcutil/build.sh -j$(nproc)
RUN cp ./zcutil/fetch-params.sh /usr/local/bin
RUN cp ./src/zclassicd /usr/local/bin
RUN cp ./src/zclassic-cli /usr/local/bin


FROM debian:stable-slim
COPY --from=build-env /sources/zclassic/src/zclassicd /usr/local/bin
COPY --from=build-env /sources/zclassic/src/zclassic-cli /usr/local/bin
COPY --from=build-env /sources/zclassic/zcutil/fetch-params.sh /usr/local/bin
EXPOSE 8232
ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/fetch-params.sh
RUN mkdir /root/.zclassic
RUN mkdir /root/.zcash-params
RUN apt-get update && apt-get -y install libgomp1

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
