package workload

import (
	"log"
	"reflect"
)

//Global variable
var support_distributions []string = append(make([]string, 0), "Uniform")

func NotSupport(distribution string) bool {
	for _, support_distribution := range support_distributions {
		if distribution == support_distribution {
			return false
		}
	}
	return true
}

func CheckWorkloadValidty(workload map[string]interface{}) bool {
	var fields_to_check map[string]interface{}
	fields_to_check["test_name"] = ""
	fields_to_check["blocking_cli"] = false
	fields_to_check["duration"] = 0
	for field, t := range fields_to_check {
		if value, ok := workload[field]; ok {
			if reflect.TypeOf(value) != reflect.TypeOf(t) {
				log.Print(field, "type should be", reflect.TypeOf(t))
				return false
			}
		} else {
			log.Println(field, "not found")
			return false
		}
	}
	if workload["duration"].(int) < 0 {
		log.Println("duration invalid")
		return false
	}
	if _, ok := workload["instances"]; !ok {
		log.Println("instances not found")
		return false
	}
	instances, ok := workload["instances"].(map[string]interface{})
	if !ok {
		log.Println("instances invalid")
		return false
	}
	for instance, desc := range instances {
		desc := desc.(map[string]interface{})
		if application, ok := desc["application"]; ok {
			if _, ok := application.(string); !ok {
				log.Println("In", instance, "application should be string")
				return false
			}
		} else {
			log.Println("In", instance, "application not found")
			return false
		}
		if distribution, ok := desc["distribution"]; ok {
			if _, ok := distribution.(string); !ok {
				log.Println("In", instance, "distribution should be string")
				return false
			}
			if NotSupport(distribution.(string)) {
				log.Println("In", instance, "not support", distribution, "distribution")
				log.Println("Support only ", support_distributions)
				return false
			}
		} else {
			log.Println("In", instance, "distribution not found")
			return false
		}
		if rate, ok := desc["rate"]; ok {
			if _, ok := rate.(int); !ok {
				log.Println("In", instance, "rate should be positive integer")
				return false
			}
			if rate.(int) < 0 {
				log.Println("In", instance, "rate should be positive integer")
				return false
			}
		} else {
			log.Println("In", instance, "rate not found")
			return false
		}
	}
	return true
}
