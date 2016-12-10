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

#set listening and broadcast address to pod ip
NODE_IP="$(hostname --ip-address)"
CASSANDRA_LISTEN_ADDRESS=$NODE_IP
CASSANDRA_BROADCAST_ADDRESS=$NODE_IP
CASSANDRA_RPC_ADDRESS=0.0.0.0

#set broadcast rpc address to avi dns
BROADCAST_IP=""
while [  "$BROADCAST_IP" = "" ]; do
	BROADCAST_IP=`dig ${CASSANDRA_NODE_NAME}.${AVI_DOMAIN_NAME} @${AVI_DNS_VIP} +short`
	sleep 5
done
CASSANDRA_BROADCAST_RPC_ADDRESS=$BROADCAST_IP

#set default options for container
CASSANDRA_AUTHENTICATOR="AllowAllAuthenticator"
CASSANDRA_AUTHORIZER="AllowAllAuthorizer"
CASSANDRA_AUTO_SNAPSHOT="false"
CASSANDRA_COMPACTION_LARGE_PARTITION_WARNING_THRESHOLD_MB=100
CASSANDRA_COMPACTION_THROUGHPUT_MB_PER_SEC=0
CASSANDRA_CONCURRENT_COMPACTORS=8
CASSANDRA_CONCURRENT_READS=64
CASSANDRA_CONCURRENT_WRITES=64
CASSANDRA_GC_WARN_THRESHOLD_IN_MS=1000
CASSANDRA_HEAP_NEWSIZE="800M"
CASSANDRA_HINTED_HANDOFF_THROTTLE_IN_KB=1024
CASSANDRA_INTER_DC_STREAM_THROUGHPUT_OUTBOUND_MEGABITS_PER_SEC=200
CASSANDRA_JMX_PORT=7199
CASSANDRA_KEY_CACHE_SIZE_IN_MB=
CASSANDRA_MAX_HEAP_SIZE="8192M"
CASSANDRA_MAX_HINTS_DELIVERY_THREADS=2
CASSANDRA_MEMTABLE_ALLOCATION_TYPE=heap_buffers
CASSANDRA_MEMTABLE_CLEANUP_THRESHOLD=0.11
CASSANDRA_MEMTABLE_FLUSH_WRITERS=4
CASSANDRA_NUM_TOKENS=64
CASSANDRA_PARTITIONER="org.apache.cassandra.dht.Murmur3Partitioner"
CASSANDRA_ROW_CACHE_SAVE_PERIOD=0
CASSANDRA_ROW_CACHE_SIZE_IN_MB=0
CASSANDRA_ENDPOINT_SNITCH="GossipingPropertyFileSnitch"
CASSANDRA_STREAM_THROUGHPUT_OUTBOUND_MEGABITS_PER_SEC=4000
CASSANDRA_STREAMING_SOCKET_TIMEOUT_IN_MS=86400000
CASSANDRA_TRICKLE_FSYNC="true"

if [ "$1" = 'cassandra' ]; then

	sed -ri 's/(- seeds:).*/\1 "'"$CASSANDRA_SEEDS"'"/' "$CASSANDRA_CONFIG/cassandra.yaml"

	for yaml in \
		authenticator \
		authorizer \
		auto_snapshot \
		broadcast_address \
		broadcast_rpc_address \
		cluster_name \
		compaction_large_partition_warning_threshold_mb \
		compaction_throughput_mb_per_sec \
		concurrent_compactors \
		concurrent_reads \
		concurrent_writes \
		endpoint_snitch \
		gc_warn_threshold_in_ms \
		hinted_handoff_throttle_in_kb \
		inter_dc_stream_throughput_outbound_megabits_per_sec \
		key_cache_size_in_mb \
		listen_address \
		max_hints_delivery_threads \
		memtable_allocation_type \
		memtable_cleanup_threshold \
		memtable_flush_writers \
		num_tokens \
		partitioner \
		row_cache_save_period \
		row_cache_size_in_mb \
		rpc_address \
		seeds \
		stream_throughput_outbound_megabits_per_sec \
		streaming_socket_timeout_in_ms \
		trickle_fsync \
	; do
		var="CASSANDRA_${yaml^^}"
		val="${!var}"
		if [ "$val" ]; then
			sed -ri 's/^(# )?('"$yaml"':).*/\2 '"$val"'/' "$CASSANDRA_CONFIG/cassandra.yaml"
		fi
	done

	#set options for cassandra-env.sh
	for env in MAX_HEAP_SIZE HEAP_NEWSIZE JMX_PORT; do
		var="CASSANDRA_${env^^}"
		val="${!var}"
		if [ "$val" ]; then
			sed -ri 's/^('"$env"'=).*/\1'"\"$val\""'/' "$CASSANDRA_CONFIG/cassandra-env.sh"
		fi
	done

	#set options for cassandra-rackdc.properties
	for rackdc in dc rack; do
		var="CASSANDRA_${rackdc^^}"
		val="${!var}"
		if [ "$val" ]; then
			sed -ri 's/^('"$rackdc"'=).*/\1'"$val"'/' "$CASSANDRA_CONFIG/cassandra-rackdc.properties"
		fi
	done

	#set options for graphite reporting
	if [ "${CASSANDRA_NODE_NAME:+1}" ]; then
		sed -ri 's/^(# )?(    prefix:).*/\2 '\'"$CASSANDRA_NODE_NAME"\''/' "$CASSANDRA_CONFIG/metrics-reporter-config.yaml"
	fi

	if [ "${GRAPHITE_HOST:+1}" ]; then
		sed -i 's/graphite_host/'"$GRAPHITE_HOST"'/' "$CASSANDRA_CONFIG/metrics-reporter-config.yaml"
	fi

	#set options for opscenter
	if [ "${OPSCENTER_HOST+1}" ]; then
		sed -ri 's/^(stomp_host: ).*/\1 '\'"$OPSCENTER_HOST"\''/' "/var/lib/datastax-agent/conf/address.yaml"
		sed -i 's/127.0.0.1/${BROADCAST_IP}/g' "/var/lib/datastax-agent/conf/address.yaml"
		/etc/init.d/datastax-agent start
	fi

fi

exec "$@"
