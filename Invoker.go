package workload

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
var config ClientConfig

func post(client *Client, action string, data map[string]interface{}) {
	switch action {
	case "Get":
		client.Get(config.ClientID, data["Key"].(string))
	case "Put":
		client.Put(config.ClientID, data["Key"].(string), data["Value"].(string), 0)
	default:
		log.Fatalln("No action", action)
	}
}

func HTTPInstanceGenerator(instance string, action string, instance_time []float32) {
	//url := base_url + action
	client := NewClient(config, kvslib.NewKVS())
	if err := client.Initialize(); err != nil {
		log.Fatalln(err)
	}
	before_time := 0
	after_time := 0
	st := 0
	for _, t := range instance_time {
		st = int(1000*t - float32(after_time-before_time))
		time.Sleep(time.Duration(st) * time.Millisecond)
		before_time = int(time.Now().Nanosecond() / 1000000)
		post(client, action, data[instance].(map[string]interface{}))
		after_time = int(time.Now().Nanosecond() / 1000000)
	}
	client.Close()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("config file must be provide as first argument")
	}
	var workload map[string]interface{}
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
	if _, ok := workload["ClientID"]; !ok {
		workload["ClientID"] = "clientID1"
	}
	config.ClientID = workload["ClientID"].(string)
	config.FrontEndAddr = workload["FrontEndAddr"].(string)
	all_event, event_count := GenericEventGenerator(workload)
	for instance := range all_event {
		file, err = ioutil.ReadFile(workload[instance].(map[string]interface{})["data_file"].(string))
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
	for instance, instance_time := range all_event {
		action := workload[instance].(map[string]interface{})["application"].(string)
		go HTTPInstanceGenerator(instance, action, instance_time.([]float32))
	}
	log.Println("Total:", event_count, "event(s)")
}
