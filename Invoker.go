package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
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

func HTTPInstanceGenerator(instance string, action string, instance_time []float32) {
	//url := base_url + action
	config := ClientConfig{
		ClientID:instance,
		FrontEndAddr:FrontEndAddr,
	}
	client := NewClient(config, kvslib.NewKVS())
	if err := client.Initialize(); err != nil {
		log.Fatalln(err)
	}
	before_time := 0
	after_time := 0
	st := 0
	time_stamp := make([]time.Time, len(instance_time))
	go ReceiveResponse(client, &time_stamp)
	for index, t := range instance_time {
		st = int(1000*t - float32(after_time-before_time))
		time.Sleep(time.Duration(st) * time.Millisecond)
		before_time = int(time.Now().Nanosecond() / 1000000)
		time_stamp[index] = time.Now()
		post(client, action, data[instance].(map[string]interface{}))
		after_time = int(time.Now().Nanosecond() / 1000000)
	}
}

func ReceiveResponse(client *Client, time_stamp *[]time.Time) {
	RoundTripTimes := make([]int64, len(*time_stamp))
	Results := make([]string, len(*time_stamp))
	i := 0
	for i < len(*time_stamp) {
		result := <-client.NotifyChannel
		log.Println("Hello")
		Results[i] = *result.Result
		RoundTripTimes[i] = int64(time.Now().Sub((*time_stamp)[i])/time.Millisecond)
		i++
	}
	client.Close()
	log.Println(client.id, "Round Trip Time:", RoundTripTimes, "Result:", Results)
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
	for instance, instance_time := range all_event {
		action := workload["instances"].(map[string]interface{})[instance].(map[string]interface{})["application"].(string)
		go HTTPInstanceGenerator(instance, action, instance_time.([]float32))
	}
	time.Sleep(time.Duration(int(workload["duration"].(float64)) + 10)*time.Second)
	log.Println("Test End")
	log.Println("Total:", event_count, "event(s)")
}
