FROM cassandra:2.1

#update image and install necessary utilities
RUN apt-get update && apt-get install -y sudo htop dstat curl dnsutils

#add modified cassandra.yaml and cassandra-env.sh to container image
COPY cassandra.yaml /etc/cassandra/cassandra.yaml
COPY cassandra-env.sh /etc/cassandra/cassandra-env.sh

#disable jmx remote authentication for demo purposes
RUN sed -i 's/jmxremote.authenticate=true/jmxremote.authenticate=false/g' /etc/cassandra/cassandra-env.sh

#add datastax agent
ADD datastax-agent.deb /tmp
RUN dpkg --install /tmp/datastax-agent.deb && rm /tmp/datastax-agent.deb
ADD address.yaml /var/lib/datastax-agent/conf/address.yaml


COPY docker-entrypoint.sh /docker-entrypoint.sh
