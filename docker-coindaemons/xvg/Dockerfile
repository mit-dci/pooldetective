FROM ubuntu:latest
ARG DEBIAN_FRONTEND=noninteractive
WORKDIR /app

# Basic env setup
RUN apt-get update
RUN apt-get -y install git
RUN apt-get -y install build-essential libtool autotools-dev automake pkg-config bsdmainutils python3
RUN git clone https://github.com/vergecurrency/VERGE.git verge

# install dependecies
RUN apt-get -y install libssl-dev libevent-dev libboost-all-dev
RUN apt-get -y install zlib1g-dev libseccomp-dev libcap-dev libncap-dev
RUN apt-get -y install software-properties-common
RUN add-apt-repository ppa:bitcoin/bitcoin
RUN apt-get update
RUN apt-get -y install libdb4.8-dev libdb4.8++-dev
RUN apt-get -y install libminiupnpc-dev
RUN apt-get -y install libzmq3-dev

# switch to the freshly cloned verge project & compile it
WORKDIR /app/verge
RUN git fetch && git checkout v6.0.2

# compile verged fresh (no ui)
RUN ./autogen.sh
RUN ./configure --disable-bench --disable-tests --disable-gui-test --disable-gui
RUN make

WORKDIR /app/verge/src

EXPOSE 20102
EXPOSE 21102

ADD ./docker_entrypoint.sh /usr/local/bin/docker_entrypoint.sh
RUN chmod a+x /usr/local/bin/docker_entrypoint.sh

RUN mkdir /root/.VERGE

ENTRYPOINT ["/usr/local/bin/docker_entrypoint.sh"]
