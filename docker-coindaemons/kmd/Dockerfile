FROM ubuntu:20.04
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"

RUN apt-get update && apt-get install -y libgomp1 wget
ADD ./komodod /usr/local/bin
ADD ./komodo-cli /usr/local/bin

EXPOSE 7771 

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
ADD ./fetch-params.sh /usr/local/bin/fetch-params.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/komodod
RUN chmod a+x /usr/local/bin/komodo-cli 
RUN chmod a+x /usr/local/bin/fetch-params.sh
RUN apt-get update && apt-get install -y libbrotli1
ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
