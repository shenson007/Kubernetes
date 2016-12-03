package main

import (
	"fmt"
	//"sort"
	"net/http"
	"strconv"
	"strings"

	"github.com/mediocregopher/radix.v2/redis"
)

const (
	api_listen_port = ":4040"
	redis_server    = "localhost:6379"
)

func main() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/cassandra/rpc_address", CassandraRPCAddress)
	http.HandleFunc("/cassandra/seeds", CassandraSeeds)
	http.ListenAndServe(api_listen_port, nil)
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=0, s-maxage=0")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "Welcome!")
}

func CassandraRPCAddress(w http.ResponseWriter, r *http.Request) {
	cluster_name := r.URL.Query().Get("cluster_name")
	fmt.Printf("cluster_name: %s\n", cluster_name)

	remote_ip := strings.Split(r.RemoteAddr, ":")[0]
	fmt.Printf("remote_ip: %s\n", remote_ip)

	fmt.Fprintf(w, "%s", remote_ip)
}

func CassandraSeeds(w http.ResponseWriter, r *http.Request) {
	//pretty := r.URL.Query().Get("pretty")
	//table := r.URL.Query().Get("table")
	cluster_name := r.URL.Query().Get("cluster_name")
	node_ip := r.URL.Query().Get("node_ip")
	max_seeds := r.URL.Query().Get("max_seeds")

	if max_seeds == "" {
		max_seeds = "2"
	}

	w.Header().Set("Cache-Control", "max-age=0, s-maxage=0")
	w.Header().Set("Content-Type", "application/json")

	fmt.Printf("cluster_name: %s\n", cluster_name)
	fmt.Printf("node_ip: %s\n", node_ip)

	seeds := RedisGetCassandraSeeds(cluster_name, node_ip, max_seeds, false)

	fmt.Fprintf(w, "%s", seeds)
}

func RedisGetCassandraSeeds(cluster_name string, node_ip string, max_seeds string, called_previously bool) string {
	redis_client, err := redis.Dial("tcp", redis_server)
	if err != nil {
		fmt.Printf("REDIS_CONNECT_ERROR: %s\n", err)
	}

	redis_key := "cassandra-config:" + cluster_name + ":seeds"
	seeds, _ := redis_client.Cmd("GET", redis_key).Str()
	if err != nil {
		fmt.Printf("REDIS_GET_ERROR: %s\n", err)
		return ""
	}
	if seeds == "" {
		seeds = node_ip
		err = redis_client.Cmd("SET", redis_key, seeds).Err
		if err != nil {
			fmt.Printf("REDIS_SET_ERROR: %s\n", err)
			return ""
		}
		return seeds
	} else {
		seeds_array := strings.Split(seeds, ",")
		seed_count := len(seeds_array)
		fmt.Printf("seed count: %d\n", seed_count)
		fmt.Printf("seeds: %v\n", seeds)
		max_seeds_int, _ := strconv.Atoi(max_seeds)
		fmt.Printf("max seeds: %d\n", max_seeds_int)

		if seed_count >= max_seeds_int {
			return seeds
		}

		for _, seed_ip := range seeds_array {
			if node_ip == seed_ip {
				return seeds
			}
		}
		//if seed_count >= 1 && count < (seed_count) {
		//    seeds += ","
		//}
		seeds += "," + node_ip

		err = redis_client.Cmd("SET", redis_key, seeds).Err
		if err != nil {
			fmt.Printf("REDIS_SET_ERROR: %s\n", err)
			return ""
		}
		if seed_count < max_seeds_int {
			return RedisGetCassandraSeeds(cluster_name, node_ip, max_seeds, true)
		}
	}
	return seeds
}
