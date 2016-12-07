package main

import (
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/marpaia/graphite-golang"
)

var auth_token string
var api_gateway string
var username string
var password string

var graphite_host string
var graphite_port int

type MappedSDC struct {
	ID          string `json:"id"`
	IP          string `json:"ip"`
	LimitBwMbps int    `json:"limit_bw_mbps"`
	LimitIops   int    `json:"limit_iops"`
}

type VolumeInstancesRaw []VolumeInstanceRaw

type VolumeInstanceRaw struct {
	CreationTime  int `json:"creationTime"`
	MappedSdcInfo []struct {
		SdcID         string `json:"sdcId"`
		SdcIP         string `json:"sdcIp"`
		LimitBwInMbps int    `json:"limitBwInMbps"`
		LimitIops     int    `json:"limitIops"`
	} `json:"mappedSdcInfo"`
	MappingToAllSdcsEnabled bool        `json:"mappingToAllSdcsEnabled"`
	UseRmcache              bool        `json:"useRmcache"`
	IsObfuscated            bool        `json:"isObfuscated"`
	VolumeType              string      `json:"volumeType"`
	ConsistencyGroupID      interface{} `json:"consistencyGroupId"`
	VtreeID                 string      `json:"vtreeId"`
	AncestorVolumeID        interface{} `json:"ancestorVolumeId"`
	StoragePoolID           string      `json:"storagePoolId"`
	SizeInKb                int         `json:"sizeInKb"`
	Name                    string      `json:"name"`
	ID                      string      `json:"id"`
	Links                   []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

type VolumePerformanceRaw struct {
	NumOfMappedScsiInitiators int `json:"numOfMappedScsiInitiators"`
	NumOfMappedSdcs           int `json:"numOfMappedSdcs"`
	NumOfChildVolumes         int `json:"numOfChildVolumes"`
	UserDataWriteBwc          struct {
		NumSeconds      int `json:"numSeconds"`
		TotalWeightInKb int `json:"totalWeightInKb"`
		NumOccured      int `json:"numOccured"`
	} `json:"userDataWriteBwc"`
	NumOfDescendantVolumes int `json:"numOfDescendantVolumes"`
	UserDataReadBwc        struct {
		NumSeconds      int `json:"numSeconds"`
		TotalWeightInKb int `json:"totalWeightInKb"`
		NumOccured      int `json:"numOccured"`
	} `json:"userDataReadBwc"`
}

type VolumePerformance struct {
	ReadKBs   int `json:"read_kbs"`
	ReadIOPs  int `json:"read_iops"`
	WriteKBs  int `json:"write_kbs"`
	WriteIOPs int `json:"write_iops"`
}

type VolumeInstances []VolumeInstance

type VolumeInstance struct {
	Name          string            `json:"name"`
	ID            string            `json:"id"`
	Size          int               `json:"size"`
	StoragePoolID string            `json:"storagepool_id"`
	Thin          bool              `json:"thin"`
	MappedSDCs    []MappedSDC       `json:"mapped_sdcs"`
	Performance   VolumePerformance `json:"performance","omitempty"`
}

type Report struct {
	Volumes VolumeInstances `json:"volumes"`
}

func main() {
	api_gateway = os.Getenv("SCALEIO_GATEWAY")
	username = os.Getenv("SCALEIO_USER")
	password = os.Getenv("SCALEIO_PASSWORD")
	graphite_host = os.Getenv("GRAPHITE_HOST")
	graphite_port_string := os.Getenv("GRAPHITE_PORT")
	graphite_port, _ = strconv.Atoi(graphite_port_string)

	fmt.Printf("gateway: %s\nusername: %s\npassword: %s\ngraphite_host: %s\ngraphite_port: %d\n", api_gateway,username,password,graphite_host,graphite_port)


	if len(password) < 1 {
		return
	}

	auth_token = SIOAuthenticate()
	volumes := VolumeInstances{}

	counter := 0

	for true {
		//Get volume list and performance metrics
		volumes := GetVolumes()
		UpdateGraphite(volumes)

		if counter == 720 {
			counter = 0
			auth_token = SIOAuthenticate()
		}

		//Sleep for 30 seconds between re-run of all previous operations
		time.Sleep(5 * time.Second)

		counter++
	}

	report := Report{
		Volumes: volumes,
	}

	report_json, _ := json.MarshalIndent(report, "", "    ")
	fmt.Println(string(report_json))
}

func SIOAuthenticate() string {
	//Create base64 encoded Authorization header for ScaleIO Gateway
	authorization_encoded := b64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	authorization_header := "Basic " + authorization_encoded

	//Init/Set request_url to ScaleIO Gateway login
	request_url := api_gateway + "/api/login"

	//Ignore SSL Certificate
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	//Setup HTTP Connection Details
	http_client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", request_url, nil)

	//Set required headers for request (only Authorization for this request)
	req.Header.Set("Authorization", authorization_header)

	//Run HTTP Get Request
	resp, err := http_client.Do(req)
	if err != nil {
		fmt.Printf("HTTP_ERROR: %s\n", err)
	}

	//Get HTTP Response Body
	resp_body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	auth_token := string(resp_body)

	auth_token = strings.TrimLeft(auth_token, `"`)
	auth_token = strings.TrimRight(auth_token, `"`)

	return auth_token
}

func GetVolumes() VolumeInstances {

	//Create base64 encoded Authorization header for ScaleIO Gateway
	authorization_encoded := b64.StdEncoding.EncodeToString([]byte(":" + auth_token))
	authorization_header := "Basic " + authorization_encoded

	//Init/Set request_url to Volume List
	request_url := api_gateway + "/api/types/Volume/instances"

	//Ignore SSL Certificate
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	//Setup HTTP Connection Details
	http_client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", request_url, nil)

	//Set required headers for request (only Authorization for this request)
	req.Header.Set("Authorization", authorization_header)
	req.Header.Set("Accept", "application/json")

	//Run HTTP Get Request
	resp, err := http_client.Do(req)
	if err != nil {
		fmt.Printf("HTTP_ERROR: %s\n", err)
	}

	//Get HTTP Response Body
	resp_body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	//Initialize volume_instances_raw to become destination for parsed jSON
	volume_instances_raw := VolumeInstancesRaw{}

	//Parse JSON into accessable struct
	err = json.Unmarshal(resp_body, &volume_instances_raw)
	if err != nil {
		fmt.Printf("JSON Parsing Error: %s\n", err)
	}

	//Initialize volume_instances to hold volumes details
	volume_instances := VolumeInstances{}

	for _, volume := range volume_instances_raw {

		size_float := float64(volume.SizeInKb) / 1024 / 1024
		size_string := fmt.Sprintf("%.0f", size_float)
		size_int, _ := strconv.Atoi(size_string)

		//Create List of Mapped SDCs
		mapped_sdcs := []MappedSDC{}
		if len(volume.MappedSdcInfo) > 0 {
			for _, mapping := range volume.MappedSdcInfo {
				mapped_sdc := MappedSDC{
					ID:          mapping.SdcID,
					IP:          mapping.SdcIP,
					LimitBwMbps: mapping.LimitBwInMbps,
					LimitIops:   mapping.LimitIops,
				}
				mapped_sdcs = append(mapped_sdcs, mapped_sdc)
			}
		}

		//Check if thin provisioned
		thin := false
		if volume.VolumeType == "ThinProvisioned" {
			thin = true
		}

		//Get performance metrics if volume attached
		performance := VolumePerformance{}
		if len(mapped_sdcs) > 0 {
			performance = GetVolumePerformance(volume.ID)
		}

		volume_instance := VolumeInstance{
			ID:            volume.ID,
			Name:          volume.Name,
			Size:          size_int,
			StoragePoolID: volume.StoragePoolID,
			Thin:          thin,
			MappedSDCs:    mapped_sdcs,
			Performance:   performance}
		volume_instances = append(volume_instances, volume_instance)

		//Get Performance Metrics if attached

	}

	sort.Sort(volume_instances)

	return volume_instances
}

func GetVolumePerformance(volumeid string) VolumePerformance {
	//Create base64 encoded Authorization header for ScaleIO Gateway
	authorization_encoded := b64.StdEncoding.EncodeToString([]byte(":" + auth_token))
	authorization_header := "Basic " + authorization_encoded

	//Init/Set request_url to Volume List
	request_url := api_gateway + "/api/instances/Volume::" + volumeid + "/relationships/Statistics"

	//Ignore SSL Certificate
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	//Setup HTTP Connection Details
	http_client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", request_url, nil)

	//Set required headers for request (only Authorization for this request)
	req.Header.Set("Authorization", authorization_header)
	req.Header.Set("Accept", "application/json")

	//Run HTTP Get Request
	resp, err := http_client.Do(req)
	if err != nil {
		fmt.Printf("HTTP_ERROR: %s\n", err)
	}

	//Get HTTP Response Body
	resp_body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	//Initialize volume_performance_raw to become destination for parsed jSON
	volume_performance_raw := VolumePerformanceRaw{}

	//Parse JSON into accessable struct
	err = json.Unmarshal(resp_body, &volume_performance_raw)
	if err != nil {
		fmt.Printf("JSON Parsing Error: %s\n", err)
	}

	//Initialize volume_instances to hold volumes details
	volume_performance := VolumePerformance{
		ReadKBs:   volume_performance_raw.UserDataReadBwc.TotalWeightInKb,
		ReadIOPs:  volume_performance_raw.UserDataReadBwc.NumOccured,
		WriteKBs:  volume_performance_raw.UserDataWriteBwc.TotalWeightInKb,
		WriteIOPs: volume_performance_raw.UserDataWriteBwc.NumOccured}

	return volume_performance
}

func UpdateGraphite(volumes VolumeInstances) {
	// try to connect a graphite server
	Graphite, err := graphite.NewGraphite(graphite_host, graphite_port)

	// if you couldn't connect to graphite, use a nop
	if err != nil {
		fmt.Printf("error connecting to graphite: %v\n", err)
		Graphite = graphite.NewGraphiteNop(graphite_host, graphite_port)
	}

	for _, volume := range volumes {
		volume_stats_json, _ := json.Marshal(volume)
		fmt.Printf("%s\n", volume_stats_json)
		Graphite.SimpleSend("scaleio.stats."+volume.Name+".read_iops.rate", strconv.Itoa(volume.Performance.ReadIOPs))
		Graphite.SimpleSend("scaleio.stats."+volume.Name+".read_kbs.rate", strconv.Itoa(volume.Performance.ReadKBs))
		Graphite.SimpleSend("scaleio.stats."+volume.Name+".write_iops.rate", strconv.Itoa(volume.Performance.WriteIOPs))
		Graphite.SimpleSend("scaleio.stats."+volume.Name+".write_kbs.rate", strconv.Itoa(volume.Performance.WriteKBs))
	}
}

//Sort-related functions for type VolumeInstances
func (slice VolumeInstances) Len() int {
	return len(slice)
}
func (slice VolumeInstances) Less(i, j int) bool {
	return slice[i].Name < slice[j].Name
}
func (slice VolumeInstances) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
