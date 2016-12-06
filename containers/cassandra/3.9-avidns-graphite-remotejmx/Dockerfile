FROM cassandra:3.9

#update image and install necessary utilities
RUN apt-get update && apt-get install -y htop dstat curl dnsutils

#add modified cassandra.yaml and cassandra-env.sh to container image
COPY cassandra.yaml /etc/cassandra/cassandra.yaml
COPY cassandra-env.sh /etc/cassandra/cassandra-env.sh

#add required files for graphite metrics reporting
COPY metrics-graphite-3.1.2.jar /usr/share/cassandra/lib
COPY metrics-reporter-config.yaml /etc/cassandra/metrics-reporter-config.yaml

#disable jmx remote authentication for demo purposes
RUN sed -i 's/jmxremote.authenticate=true/jmxremote.authenticate=false/g' /etc/cassandra/cassandra-env.sh

COPY docker-entrypoint.sh /docker-entrypoint.sh
