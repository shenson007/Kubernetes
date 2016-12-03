package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/mediocregopher/radix.v2/redis"
)

const (
	api_listen_port   = ":8089"
	avi_authorization = "user:pass"
	redis_server      = "localhost:6379"
)

type ServiceDetails struct {
	Address           string  `json:"address"`
	AlertHigh         int     `json:"alert_high"`
	AlertMedium       int     `json:"alert_medium"`
	AlertLow          int     `json:"alert_low"`
	HealthScore       float64 `json:"health_score"`
	Name              string  `json:"name"`
	NumberSEAssigned  int     `json:"num_se_assigned"`
	NumberSERequested int     `json:"num_se_requested"`
	PercentageSEsUp   int     `json:"percent_ses_up"`
	State             string  `json:"state"`
	UUID              string  `json:"uuid"`
}

type ServiceDetailsList []ServiceDetails

func main() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/virtual_services", VirtualServices)
	http.ListenAndServe(api_listen_port, nil)
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=0, s-maxage=0")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "Welcome!")
}

func VirtualServices(w http.ResponseWriter, r *http.Request) {
	pretty := r.URL.Query().Get("pretty")
	table := r.URL.Query().Get("table")
	uuid := r.URL.Query().Get("uuid")

	w.Header().Set("Cache-Control", "max-age=0, s-maxage=0")
	w.Header().Set("Content-Type", "application/json")

	if uuid != "" {
		service_details_redis := AviGetVirtualService(uuid)
		if pretty == "yes" {
			service_details := ServiceDetails{}
			json.Unmarshal([]byte(service_details_redis), &service_details)
			service_details_json, _ := json.MarshalIndent(service_details, "", "    ")
			fmt.Fprintf(w, "%s", service_details_json)
		} else {
			fmt.Fprintf(w, "%s\n", service_details_redis)
		}
	} else {
		all_service_details := AviGetVirtualServices()
		if table == "yes" {
			sort.Sort(all_service_details)
			fmt.Fprintf(w, "virtual service details\n")
			mask_header := "| %-24s | %-15s | %-13s | %12s | %7s | %7s | %7s | %11s | %12s | %13s |\n"
			mask_data := "| %-24s | %-15s | %-13s | %12.0f | %7d | %7d | %7d | %11d | %12d | %13d |\n"
			fmt.Fprintf(w, mask_header, "name", "address", "state", "health_score", "alert.l", "alert.m", "alert.h", "se_assigned", "se_requested", "se_percent_up")
			for _, result := range all_service_details {
				fmt.Fprintf(w, mask_data, result.Name, result.Address, result.State, result.HealthScore, result.AlertLow, result.AlertMedium, result.AlertHigh, result.NumberSEAssigned, result.NumberSERequested, result.PercentageSEsUp)
			}
		} else {
			if pretty == "yes" {
				all_service_details_json, _ := json.MarshalIndent(all_service_details, "", "    ")
				fmt.Fprintf(w, "%s", all_service_details_json)
			} else {
				all_service_details_json, _ := json.Marshal(all_service_details)
				fmt.Fprintf(w, "%s\n", all_service_details_json)
			}
		}
	}

}

func AviGetVirtualService(uuid string) string {
	//Connect to Redis
	redis_client, err := redis.Dial("tcp", redis_server)
	if err != nil {
		fmt.Printf("REDIS_CONNECT_ERROR: %s\n", err)
	}

	redis_key := "avi:service_details:uuid:" + uuid
	service_details_redis, err := redis_client.Cmd("GET", redis_key).Str()
	if err != nil {
		fmt.Printf("REDIS_GET_UUID_ERROR: %s\n", err)
		fmt.Printf("REDIS uuid: %s\n", uuid)
	}
	return service_details_redis
}

func AviGetVirtualServices() ServiceDetailsList {
	//Connect to Redis
	redis_client, err := redis.Dial("tcp", redis_server)
	if err != nil {
		fmt.Printf("REDIS_CONNECT_ERROR: %s\n", err)
	}

	//Check Redis for Service Details
	all_service_details_redis, err := redis_client.Cmd("GET", "avi:service_details:all").Str()
	if err != nil {
		//fmt.Printf("error: %s\n", err)
	}

	all_service_details := ServiceDetailsList{}

	json.Unmarshal([]byte(all_service_details_redis), &all_service_details)

	return all_service_details
}

//Sort-related functions for type ServiceDetailsList
func (slice ServiceDetailsList) Len() int {
	return len(slice)
}
func (slice ServiceDetailsList) Less(i, j int) bool {
	return slice[i].Name < slice[j].Name
}
func (slice ServiceDetailsList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
