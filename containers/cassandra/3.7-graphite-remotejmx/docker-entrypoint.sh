#!/bin/bash
set -e

# first arg is `-f` or `--some-option`
if [ "${1:0:1}" = '-' ]; then
	set -- cassandra -f "$@"
fi

# allow the container to be started with `--user`
if [ "$1" = 'cassandra' -a "$(id -u)" = '0' ]; then
	chown -R cassandra /var/lib/cassandra /var/log/cassandra "$CASSANDRA_CONFIG"
	exec gosu cassandra "$BASH_SOURCE" "$@"
fi

#CASSANDRA_SEEDS=`dig $CASSANDRA_SEED +short`

BROADCAST_IP=""

while [  "$BROADCAST_IP" = "" ]; do
	BROADCAST_IP=`dig ${CASSANDRA_NODE_NAME}.${AVI_DOMAIN_NAME} @${AVI_DNS_VIP} +short`
	sleep 5
done

NODE_IP="$(hostname --ip-address)"
CASSANDRA_LISTEN_ADDRESS=$NODE_IP
CASSANDRA_BROADCAST_ADDRESS=$NODE_IP
CASSANDRA_BROADCAST_RPC_ADDRESS=$BROADCAST_IP
CASSANDRA_RPC_ADDRESS=0.0.0.0

if [ "$1" = 'cassandra' ]; then

	sed -ri 's/(- seeds:).*/\1 "'"$CASSANDRA_SEEDS"'"/' "$CASSANDRA_CONFIG/cassandra.yaml"

	for yaml in \
		broadcast_address \
		broadcast_rpc_address \
		cluster_name \
		endpoint_snitch \
		listen_address \
		num_tokens \
		rpc_address \
		start_rpc \
		storage_port \
        native_transport_port \
	; do
		var="CASSANDRA_${yaml^^}"
		val="${!var}"
		if [ "$val" ]; then
			sed -ri 's/^(# )?('"$yaml"':).*/\2 '"$val"'/' "$CASSANDRA_CONFIG/cassandra.yaml"
		fi
	done

	for rackdc in dc rack; do
		var="CASSANDRA_${rackdc^^}"
		val="${!var}"
		if [ "$val" ]; then
			sed -ri 's/^('"$rackdc"'=).*/\1'"$val"'/' "$CASSANDRA_CONFIG/cassandra-rackdc.properties"
		fi
	done

        if [ "CASSANDRA_NODE_NAME" ]; then
                sed -ri 's/^(# )?(    prefix:).*/\2 '\'"$CASSANDRA_NODE_NAME"\''/' "$CASSANDRA_CONFIG/metrics-reporter-config.yaml"
        fi

        if [ ${GRAPHITE_HOST:+1} ]; then
		sed -i 's/graphite_host/'"$GRAPHITE_HOST"'/' "$CASSANDRA_CONFIG/metrics-reporter-config.yaml"
        fi
fi

exec "$@"
