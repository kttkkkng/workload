package workload

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var data map[string][]byte
var base_url string

func post(url string, body []byte) {
	//Unimplemented
}

func HTTPInstanceGenerator(instance string, action string, instance_time []float32, blocking_cli bool) {
	url := base_url + action
	before_time := 0
	after_time := 0
	st := 0
	for _, t := range instance_time {
		st = int(1000*t - float32(after_time-before_time))
		time.Sleep(time.Duration(st) * time.Millisecond)
		before_time = int(time.Now().Nanosecond() / 1000000)
		post(url, data[instance])
		after_time = int(time.Now().Nanosecond() / 1000000)
	}
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
	all_event, event_count := GenericEventGenerator(workload)
	for instance := range all_event {
		if path, ok := workload[instance].(map[string]interface{})["data_file"]; ok {
			file, err = ioutil.ReadFile(path.(string))
			if err != nil {
				log.Fatalln(err)
			}
			data[instance] = file
		} else {
			data[instance] = []byte("{}")
		}
	}
	for instance, instance_time := range all_event {
		action := workload[instance].(map[string]interface{})["application"].(string)
		blocking_cli := workload[instance].(map[string]interface{})["blocking_cli"].(bool)
		go HTTPInstanceGenerator(instance, action, instance_time.([]float32), blocking_cli)
	}
	log.Println("Total:", event_count, "event(s)")
}
