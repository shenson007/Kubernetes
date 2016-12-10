FROM cassandra:2.2

#update image and install necessary utilities
RUN apt-get update && apt-get install -y sudo htop dstat curl dnsutils sysstat

#add modified cassandra.yaml and cassandra-env.sh to container image
COPY cassandra.yaml /etc/cassandra/cassandra.yaml
COPY cassandra-env.sh /etc/cassandra/cassandra-env.sh

#add required files for graphite metrics reporting
COPY metrics-graphite-3.1.2.jar /usr/share/cassandra/lib
COPY metrics-reporter-config.yaml /etc/cassandra/metrics-reporter-config.yaml

#disable jmx remote authentication for demo purposes
RUN sed -i 's/jmxremote.authenticate=true/jmxremote.authenticate=false/g' /etc/cassandra/cassandra-env.sh

#add datastax agent
ADD datastax-agent.deb /tmp
RUN dpkg --install /tmp/datastax-agent.deb && rm /tmp/datastax-agent.deb
ADD address.yaml /var/lib/datastax-agent/conf/address.yaml


COPY docker-entrypoint.sh /docker-entrypoint.sh
