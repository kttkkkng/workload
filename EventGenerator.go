package main

import (
	"log"
)

//create waiting period, the time that go routine have to wait before send the next request
//add distribution in this func
func CreateEvents(distribution string, rate int, duration int) []float32 {
	if distribution == "Uniform" {
		n := duration * rate
		times := make([]float32, n)
		i := 0
		for i < n {
			times[i] = 1.0 / float32(rate)
			i++
		}
		return times
	}
	return make([]float32, 0)
}

//limit the activity of the application
func EnforceActivityWindow(start_time int, end_time int, times []float32) []float32 {
	n := len(times)
	i := 1
	out_of_rage := 0
	for i < n {
		times[i] = times[i] + times[i-1]
		if times[i] < float32(start_time) || times[i] > float32(end_time) {
			out_of_rage++
		}
		i++
	}
	event_times := make([]float32, n-out_of_rage)
	i = 0
	j := 0
	for i < n {
		if times[i] >= float32(start_time) && times[i] <= float32(end_time) {
			event_times[j] = times[i]
			j++
		}
		i++
	}
	i = len(event_times) - 1
	for i >= 1 {
		event_times[i] -= event_times[i-1]
		i--
	}
	if event_times[len(event_times)-1] <= 0 {
		event_times = event_times[:len(event_times)-1]
	}

	//the waiting time period(second) of all request
	return event_times
}

//create schedule of all of the sent request go routine
func GenericEventGenerator(workload map[string]interface{}) (map[string]interface{}, int) {
	duration := workload["duration"]
	all_event := make(map[string]interface{})
	event_count := 0
	for instance, value := range workload["instances"].(map[string]interface{}) {
		log.Println("Generate", instance)
		desc := value.(map[string]interface{})
		instance_events := CreateEvents(desc["distribution"].(string), int(desc["rate"].(float64)), int(duration.(float64)))
		log.Println(instance, "is created")
		start_time := 0
		end_time := int(duration.(float64))
		if activity_window, ok := desc["activity_window"]; ok {
			if window, ok := activity_window.([]interface{}); ok {
				if len(window) >= 2 {
					start_time = int(window[0].(float64))
					end_time = int(window[1].(float64))
				}
			}
		}
		instance_events = EnforceActivityWindow(start_time, end_time, instance_events)
		all_event[instance] = instance_events
		event_count += len(instance_events)
	}
	return all_event, event_count
}
