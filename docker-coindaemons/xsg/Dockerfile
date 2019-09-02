## Original from: https://github.com/jamesob/docker-bitcoind
FROM ubuntu:18.04
LABEL MAINTAINER="Gert-Jaap Glasbergen <gertjaap@mit.edu>"


# Some of this was unabashadly yanked from
# https://github.com/szyhf/DIDockerfiles/blob/master/bitcoin/alpine/Dockerfile

RUN apt-get update && apt-get install -y libzmq3-dev libboost-all-dev
RUN apt-get install -y software-properties-common
RUN apt-get install -y build-essential libssl-dev libboost-all-dev libqrencode-dev libminiupnpc-dev
RUN add-apt-repository -y ppa:bitcoin/bitcoin
RUN apt-get update
RUN apt-get install -y libdb4.8-dev libdb4.8++-dev
RUN apt-get install -y libevent-pthreads-2.1-6 libevent-2.1-6
RUN apt-get install -y wget 
RUN wget http://archive.ubuntu.com/ubuntu/pool/main/g/gcc-10/gcc-10-base_10-20200411-0ubuntu1_amd64.deb
RUN wget http://archive.ubuntu.com/ubuntu/pool/main/g/gcc-10/libgcc-s1_10-20200411-0ubuntu1_amd64.deb
RUN wget http://archive.ubuntu.com/ubuntu/pool/main/g/gcc-10/libstdc++6_10-20200411-0ubuntu1_amd64.deb && dpkg -i *.deb && apt install --reinstall libstdc++6
RUN apt-get install -y libssl1.0.0
RUN mkdir /root/.snowgem
COPY snowgem* /usr/local/bin/

EXPOSE 16112


ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
ADD ./fetch-params.sh /usr/local/bin/fetch-params.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/fetch-params.sh 

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
