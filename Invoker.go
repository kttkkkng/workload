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

//send request
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

//create new client and send request every time period
func HTTPInstanceGenerator(wg *sync.WaitGroup, instance string, action string, instance_time []float32) {
	//tell main function that this go routine is done
	defer wg.Done()
	//url := base_url + action

	//KVS Client
	config := ClientConfig{
		ClientID:     instance,
		FrontEndAddr: FrontEndAddr,
	}
	client := NewClient(config, kvslib.NewKVS())
	if err := client.Initialize(); err != nil {
		log.Fatalln(err)
	}
	//KVS Client

	stamp := time.Now()
	var time_diff time.Duration
	time_diff = 0
	st := 0
	time_stamp := make([]time.Time, len(instance_time))

	finished := make(chan bool)
	go ReceiveResponse(finished, client, &time_stamp)

	for index, t := range instance_time {
		//the time that go routine has to sleep util sending next request in microsecond
		st = int(1000000*t - float32(time_diff/time.Microsecond))

		time.Sleep(time.Duration(st) * time.Microsecond)

		stamp = time.Now()
		//request-sent time stamp
		time_stamp[index] = time.Now()
		post(uint32(index), client, action, data[instance].(map[string]interface{}))
		time_diff = time.Since(stamp)
	}

	//wait util receive all response
	<-finished
}

//receive all response and then print result of every request
func ReceiveResponse(finished chan bool, client *Client, time_stamp *[]time.Time) {
	RoundTripTimes := make([]int64, len(*time_stamp))
	Results := make([]string, len(*time_stamp))
	timeout := 0
	fail := 0
	i := 0
	for ; i < len(*time_stamp); i++ {
		result := <-client.NotifyChannel

		//duration since request was sent util receiving response in millisecond
		//maybe have to change to microsecond if the application response in 1 millisecond
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
	// Result Print Section
	cnt := len(RoundTripTimes)
	log.Println(client.id, "Round Trip Time:", RoundTripTimes, "time out:", timeout, "failed:", fail, "success:", cnt-(fail+timeout), "total:", cnt)
	// End Result Print Section
	finished <- true
}

//read config from path that provide as first argument
//check if config is valid by call CheckWorkloadValidty
//generate request schedule by call GenericEventGenerator
//read data file of every instance
//start the test, every instance run parallel
//test end if receive all response
func main() {
	if len(os.Args) < 2 {
		log.Fatalln("config file must be provide as first argument")
	}

	var workload map[string]interface{}
	data = make(map[string]interface{})

	//read config file and check validty
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
	FrontEndAddr = workload["FrontEndAddr"].(string)

	all_event, event_count := GenericEventGenerator(workload)
	log.Println("GenericEventGen Completed")

	//read all data file
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

	if _, ok := workload["test_name"]; !ok {
		workload["test_name"] = ""
	}
	if name, ok := workload["test_name"].(string); ok {
		log.Println("Test", name, "Start")
	} else {
		log.Println("Test Start")
	}

	var wg sync.WaitGroup

	//start every instance
	for instance, instance_time := range all_event {
		action := workload["instances"].(map[string]interface{})[instance].(map[string]interface{})["application"].(string)
		wg.Add(1)
		go HTTPInstanceGenerator(&wg, instance, action, instance_time.([]float32))
	}

	//wait util every instance end
	wg.Wait()

	log.Println("Test End")
	log.Println("Total:", event_count, "event(s)")
}
