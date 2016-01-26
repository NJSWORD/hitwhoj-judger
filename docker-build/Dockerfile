FROM ubuntu:14.04
ADD sources.list /etc/apt/
RUN apt-get update && apt-get -y install build-essential openjdk-7-jdk git rake pkg-config make autoconf automake libtool libseccomp-dev
RUN git clone https://github.com/quark-zju/lrun
RUN cd lrun && make install

RUN useradd -m -s /bin/bash judge


ADD judger /usr/bin/

CMD judger -redis-host=$REDIS_PORT_6379_TCP_ADDR -redis-port=$REDIS_PORT_6379_TCP_PORT -mongo-host=$MONGO_PORT_27017_TCP_ADDR -mongo-port=$MONGO_PORT_27017_TCP_PORT
