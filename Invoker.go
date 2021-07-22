package workload

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

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
}
