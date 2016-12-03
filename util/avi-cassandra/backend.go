package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
)

const (
	api_public_domain = "api.domain.com"
	api_public_port   = "443"
	avi_authorization = "admin:Avi12345"
	avi_request_url   = "http://avi-controller/api/virtualservice-inventory/?limit=1"
	redis_default_ttl = 600
	redis_server      = "localhost:6379"
)

type AviServiceDetails struct {
	Address           string  `json:"address"`
	AlertHigh         int     `json:"alert_high"`
	AlertMedium       int     `json:"alert_medium"`
	AlertLow          int     `json:"alert_low"`
	HealthScore       float64 `json:"health_score"`
	Name              string  `json:"name"`
	NumberSeAssigned  int     `json:"num_se_assigned"`
	NumberSeRequested int     `json:"num_se_requested"`
	PercentageSesUp   int     `json:"percent_ses_up"`
	State             string  `json:"state"`
	Uuid              string  `json:"uuid"`
}

type AviServiceDetailsList []AviServiceDetails

type AviAlerts struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
}

type AviHealthScores struct {
	HealthScore float64 `json:"health_score"`
}

type AviOperStatus struct {
	State string `json:"state"`
}

type AviRuntime struct {
	OperStatus        AviOperStatus `json:"oper_status"`
	NumberSeAssigned  int           `json:"num_se_assigned"`
	NumberSeRequested int           `json:"num_se_requested"`
	PercentageSesUp   int           `json:"percent_ses_up"`
}

type AviIpAddress struct {
	Address string `json:"addr"`
}

type AviConfig struct {
	Address   string       `json:"address"`
	Name      string       `json:"name"`
	IpAddress AviIpAddress `json:"ip_address"`
}

type AviVipDetails struct {
	Alert       AviAlerts       `json:"alert"`
	HealthScore AviHealthScores `json:"health_score"`
	Runtime     AviRuntime      `json:"runtime"`
	Config      AviConfig       `json:"config"`
	Uuid        string          `json:"uuid"`
}

type AviVirtualServiceInventoryResponse struct {
	Results []AviVipDetails `json:"results"`
}

func main() {
	for true {
		//Get Most Recent Config Details from both Avi and Circonus APIs
		//avi_virtual_services := AviGetVirtualServices()
		AviGetVirtualServices()
		//Sleep for 5 seconds between re-run of all previous operations
		time.Sleep(5 * time.Second)

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

func AviGetVirtualServices() AviServiceDetailsList {
	//Print notice to terminal that function has been called
	fmt.Printf("Getting Virtual Services\n")

	//Initialize function's return variable
	all_service_details := AviServiceDetailsList{}

	//Create base64 encoded Authorization header for Avi API
	authorization_encoded := b64.StdEncoding.EncodeToString([]byte(avi_authorization))
	authorization_encoded = "Basic " + authorization_encoded

	//Init/Set request_url to Avi API request
	request_url := avi_request_url

	//Initialize and Configure a new HTTP Get Request
	http_client := &http.Client{}
	req, err := http.NewRequest("GET", request_url, nil)

	//Set required headers for request (only Authorization for this request)
	req.Header.Set("Authorization", authorization_encoded)

	//Set Service to default
	req.Header.Set("X-Avi-Tenant", "default")

	//Run HTTP Get Request
	resp, err := http_client.Do(req)
	if err != nil {
		fmt.Printf("HTTP_ERROR: %s\n", err)
	}

	//Get HTTP Response Body
	resp_body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	//Initialize avi_response to become destination for parsed jSON
	avi_response := AviVirtualServiceInventoryResponse{}

	//Parse JSON into accessable struct
	err = json.Unmarshal(resp_body, &avi_response)
	if err != nil {
		fmt.Printf("JSON Parsing Error: %s\n", err)
	}

	//Foreach service in parsed JSON, gather useful data and copy to end of previously declared all_service_details return struct
	for _, service := range avi_response.Results {
		temp := AviServiceDetails{}
		temp.Address = service.Config.IpAddress.Address
		temp.AlertHigh = service.Alert.High
		temp.AlertMedium = service.Alert.Medium
		temp.AlertLow = service.Alert.Low
		temp.HealthScore = service.HealthScore.HealthScore
		temp.Name = service.Config.Name
		temp.NumberSeAssigned = service.Runtime.NumberSeAssigned
		temp.NumberSeRequested = service.Runtime.NumberSeRequested
		temp.PercentageSesUp = service.Runtime.PercentageSesUp
		temp.State = service.Runtime.OperStatus.State
		temp.Uuid = service.Uuid

		all_service_details = append(all_service_details, temp)
	}

	//Connect to Redis
	redis_client, err := redis.Dial("tcp", redis_server)
	if err != nil {
		fmt.Printf("REDIS_CONNECT_ERROR: %s\n", err)
	}

	//Create JSON String for all_service_details and Store in Redis
	all_service_details_json, _ := json.Marshal(all_service_details)
	err = redis_client.Cmd("SET", "avi:service_details:all", all_service_details_json, "EX", redis_default_ttl).Err
	if err != nil {
		fmt.Printf("REDIS_SET_ERROR: %s\n", err)
	}

	//Cycle through all services, Create JSON strings, and store in Redis
	for _, service_details := range all_service_details {
		//Set Redis Key for location of individual vip details JSON
		redis_key := "avi:service_details:uuid:" + service_details.Uuid

		//Convert service_details struct to single-line JSON
		service_details_json, _ := json.Marshal(service_details)

		//Copy single-line JSON to Redis with TTL of redis_default_ttl
		err = redis_client.Cmd("SET", redis_key, service_details_json, "EX", redis_default_ttl).Err
		if err != nil {
			fmt.Printf("REDIS_SET_ERROR: %s\n", err)
		}

		redis_key = "avi:ip_lookup:" + service_details.Name

		err = redis_client.Cmd("SET", redis_key, service_details.Address, "EX", redis_default_ttl).Err
		if err != nil {
			fmt.Printf("REDIS_SET_ERROR: %s\n", err)
		}

	}

	return all_service_details
}
