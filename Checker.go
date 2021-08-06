package main

import (
	"log"
	"os"
)

//Global variable
var support_distributions []string = []string{"Uniform"}

//Check if the distribution is supported distribution
func NotSupport(distribution string) bool {
	for _, support_distribution := range support_distributions {
		if distribution == support_distribution {
			return false
		}
	}
	return true
}

//Check if the json is valid
//test_name: the name of this test
//duration: the duration of this test
//FronEndAddr: address of the frontend service (must reimplement this when change the application)
//instances: detail of every instaances which a instance count as 1 go routine that send request with a specific rate, application, distribution, activity_window and data
//rate: in Uniform distribution mean the amount of request per second
//application: the function to invoke, e.g. Get, Put (must reimplement this when change the application)
//distribution: statistic distribution of the request (must add another distribution like Poisson distribution)
//activity_window: the start time and end time of instance
//dataFile: file path of the data which is used to send the request
func CheckWorkloadValidty(workload map[string]interface{}) bool {
	if _, ok := workload["test_name"]; !ok {
		log.Fatalln("test name not found")
	}
	if _, ok := workload["duration"]; !ok {
		log.Fatalln("duration not found")
	}
	if workload["duration"].(float64) < 0 {
		log.Println("duration should be positive integer")
		return false
	}
	if _, ok := workload["FrontEndAddr"]; !ok {
		log.Println("FrontEndAddr not found")
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
			if _, ok := rate.(float64); !ok {
				log.Println("In", instance, "rate should be positive integer")
				return false
			}
			if rate.(float64) < 0 {
				log.Println("In", instance, "rate should be positive integer")
				return false
			}
		} else {
			log.Println("In", instance, "rate not found")
			return false
		}
		if dataFile, ok := desc["dataFile"]; ok {
			if _, err := os.Stat(dataFile.(string)); os.IsNotExist(err) {
				log.Fatalln("In", instance, "data file is not exist")
			}
		} else {
			log.Println("In", instance, "data file not found")
			return false
		}
	}
	return true
}
