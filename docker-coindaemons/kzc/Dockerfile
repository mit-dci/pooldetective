FROM ubuntu:latest
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get -y install wget 
WORKDIR /tmp
RUN wget https://github.com/kzcashteam/kzcash/releases/download/v0.1.9.1/kzcash-0.1.9.1-cli-linux-ubuntu1604.tar.gz
RUN tar -xvf kzcash-0.1.9.1-cli-linux-ubuntu1604.tar.gz
RUN cp /tmp/kzc* /usr/local/bin

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/kzcashd
RUN chmod a+x /usr/local/bin/kzcash-cli

RUN mkdir /root/.kzcash

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]