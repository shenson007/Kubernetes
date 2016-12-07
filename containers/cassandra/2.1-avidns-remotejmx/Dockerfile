FROM cassandra:2.1

#update image and install necessary utilities
RUN apt-get update && apt-get install -y htop dstat curl dnsutils

#add modified cassandra.yaml and cassandra-env.sh to container image
COPY cassandra.yaml /etc/cassandra/cassandra.yaml
COPY cassandra-env.sh /etc/cassandra/cassandra-env.sh

#disable jmx remote authentication for demo purposes
RUN sed -i 's/jmxremote.authenticate=true/jmxremote.authenticate=false/g' /etc/cassandra/cassandra-env.sh

COPY docker-entrypoint.sh /docker-entrypoint.sh
