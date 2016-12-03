FROM cassandra:3.7
COPY cassandra.yaml /etc/cassandra/cassandra.yaml

RUN apt-get update && apt-get install -y htop dstat curl dnsutils

COPY metrics-graphite-3.1.2.jar /usr/share/cassandra/lib
COPY metrics-reporter-config.yaml /etc/cassandra/metrics-reporter-config.yaml

COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN sed -i 's/jmxremote.authenticate=true/jmxremote.authenticate=false/g' /etc/cassandra/cassandra-env.sh
