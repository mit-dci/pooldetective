FROM ubuntu:latest
RUN apt-get update && apt-get -y install wget 
WORKDIR /tmp
RUN wget https://github.com/velescore/veles/releases/download/v0.18.1.3/veles-0.18.1.3-debian-ubuntu-amd64.tar.gz
RUN tar -xvf veles-0.18.1.3-debian-ubuntu-amd64.tar.gz
WORKDIR /tmp/veles-debian-ubuntu-amd64 
RUN dpkg -i velesd.deb

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/velesd

RUN mkdir /root/.velescore

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]