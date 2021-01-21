FROM debian:stable-slim
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y libgomp1 unzip wget
RUN mkdir /download && cd /download && wget https://github.com/btcz/bitcoinz/releases/download/2.0.6/bitcoinz-2.0.6-2235d5781-linux64.zip && unzip bitcoinz-2.0.6-2235d5781-linux64.zip \
  && rm *.zip && mv * /usr/local/bin/ 

EXPOSE 1979 

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/fetch-params.sh
RUN chmod a+x /usr/local/bin/bitcoinzd
RUN chmod a+x /usr/local/bin/bitcoinz-cli
ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
