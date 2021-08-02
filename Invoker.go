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

func post(reqId uint32, client *Client, action string, data map[string]interface{}) {
	switch action {
	case "Get":
		go client.Get(reqId, data["Key"].(string))
	case "Put":
		go client.Put(reqId, data["Key"].(string), data["Value"].(string), 0)
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
		post(uint32(index), client, action, data[instance].(map[string]interface{}))
		time_diff = time.Since(stamp)
	}
	<-finished
}

func ReceiveResponse(finished chan bool, client *Client, time_stamp *[]time.Time) {
	RoundTripTimes := make([]int64, len(*time_stamp))
	Results := make([]string, len(*time_stamp))
	timeout := 0
	fail := 0
	for i := 0; i < len(*time_stamp); i++ {
		result := <-client.NotifyChannel
		RoundTripTimes[result.ReqId] = int64(time.Since((*time_stamp)[result.ReqId]) / time.Millisecond)
		if result.Timeout {
			timeout++
		} else if result.StorageFail {
			fail++
		} else {
			Results[result.ReqId] = *result.Result
		}
	}
	client.Close()
	log.Println(client.id, "time out:", timeout, "failed:", fail, "Round Trip Time:", RoundTripTimes)
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
