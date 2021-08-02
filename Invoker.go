package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/kttkkkng/workload/kvslib"
)

var data map[string]interface{}
var base_url string
var FrontEndAddr string

func post(client *Client, action string, data map[string]interface{}) {
	switch action {
	case "Get":
		client.Get(client.id, data["Key"].(string))
	case "Put":
		client.Put(client.id, data["Key"].(string), data["Value"].(string), 0)
	default:
		log.Fatalln("No action", action)
	}
}

func HTTPInstanceGenerator(wg *sync.WaitGroup, instance string, action string, instance_time []float32) {
	defer wg.Done()
	//url := base_url + action
	config := ClientConfig{
		ClientID:     instance,
		FrontEndAddr: FrontEndAddr,
	}
	client := NewClient(config, kvslib.NewKVS())
	if err := client.Initialize(); err != nil {
		log.Fatalln(err)
	}
	stamp := time.Now()
	var time_diff time.Duration
	time_diff = 0
	st := 0
	time_stamp := make([]time.Time, len(instance_time))
	finished := make(chan bool)
	go ReceiveResponse(finished, client, &time_stamp)
	for index, t := range instance_time {
		st = int(1000*t - float32(time_diff/time.Millisecond))
		time.Sleep(time.Duration(st) * time.Millisecond)
		stamp = time.Now()
		time_stamp[index] = time.Now()
		post(client, action, data[instance].(map[string]interface{}))
		time_diff = time.Since(stamp)
	}
	<-finished
}

func ReceiveResponse(finished chan bool, client *Client, time_stamp *[]time.Time) {
	RoundTripTimes := make([]int64, len(*time_stamp))
	Results := make([]string, len(*time_stamp))
	i := 0
	for i < len(*time_stamp) {
		result := <-client.NotifyChannel
		if int(result.OpId) > i+1 {
			j := i + 1
			for j < int(result.OpId) {
				log.Println(client.id, "Result", j, "missing")
				RoundTripTimes[j-1] = -1
				j++
				i++
			}
		}
		Results[i] = *result.Result
		RoundTripTimes[i] = int64(time.Since((*time_stamp)[i]) / time.Millisecond)
		i++
	}
	client.Close()
	log.Println(client.id, "Round Trip Time:", RoundTripTimes)
	finished <- true
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("config file must be provide as first argument")
	}
	var workload map[string]interface{}
	data = make(map[string]interface{})
	file, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal([]byte(file), &workload)
	if err != nil {
		log.Fatalln(err)
	}
	if !CheckWorkloadValidty(workload) {
		log.Fatalln("json not valid")
	}
	log.Println("Config file is valid")
	if _, ok := workload["ClientID"]; !ok {
		workload["ClientID"] = "clientID1"
	}
	FrontEndAddr = workload["FrontEndAddr"].(string)
	all_event, event_count := GenericEventGenerator(workload)
	log.Println("GenericEventGen Completed")
	for instance := range all_event {
		file, err = ioutil.ReadFile(workload["instances"].(map[string]interface{})[instance].(map[string]interface{})["dataFile"].(string))
		if err != nil {
			log.Fatalln(err)
		}
		var tmp map[string]interface{}
		err = json.Unmarshal([]byte(file), &tmp)
		if err != nil {
			log.Fatalln(err)
		}
		data[instance] = tmp
	}
	log.Println("Test", workload["test_name"].(string), "Start")
	var wg sync.WaitGroup
	for instance, instance_time := range all_event {
		action := workload["instances"].(map[string]interface{})[instance].(map[string]interface{})["application"].(string)
		wg.Add(1)
		go HTTPInstanceGenerator(&wg, instance, action, instance_time.([]float32))
	}
	wg.Wait()
	log.Println("Test End")
	log.Println("Total:", event_count, "event(s)")
}
