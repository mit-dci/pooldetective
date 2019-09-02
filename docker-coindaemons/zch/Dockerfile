## Original from: https://github.com/jamesob/docker-bitcoind
FROM debian:stable-slim 
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get -y install wget libgomp1
RUN wget https://github.com/0cash/0cash/releases/download/v1.0.0-fix/0cash-1.0.0-amd64.deb
RUN dpkg -i 0cash-1.0.0-amd64.deb

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

RUN mkdir /root/.0cash

ADD ./fetch-params.sh /usr/local/bin/fetch-params.sh
RUN chmod a+x /usr/local/bin/fetch-params.sh
ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
