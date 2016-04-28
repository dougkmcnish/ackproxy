package main

import (
	"encoding/json"
	"fmt"
	"github.com/easy-bot/ackproxy/response"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var queue []response.Ack

var akey string = os.Getenv("NAGIOS_API_KEY")
var host string = os.Getenv("NAGIOS_API_HOST")

func main() {

	if host == "" {
		log.Fatal("Specify a host.")
	}

	if len(akey) < 30 {
		log.Fatal("Must set NAGIOS_API_KEY to something pretty secure")
	}

	res, err := http.Get("http://" + host + "/dequeue?k=" + akey)

	if err != nil {
		log.Fatalf("Connection error %s", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		log.Fatalf("Read Error: %s", err)
	}

	json.Unmarshal(body, &queue)

	if queue != nil {
		for _, a := range queue {

			time := time.Now().Unix()

			if a.Key == akey {
				if a.Service == "" {
					fmt.Printf("[%d] ACKNOWLEDGE_HOST_PROBLEM;%s;2;1;1;%s;Acknowledged using nagproxy.go\n", time, a.Host, a.User)
				} else {
					fmt.Printf("[%d] ACKNOWLEDGE_SVC_PROBLEM;%s;%s;2;1;1;%s;Acknowledged using nagproxy.go\n", time, a.Host, a.Service, a.User)
				}
			}

		}
	}

}
