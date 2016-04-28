package main

import (
	"encoding/json"
	"fmt"
	"github.com/easy-bot/ackproxy/response"
	"log"
	"net/http"
	"os"
)

var queue []response.Ack
var akey string = os.Getenv("NAGIOS_API_KEY")

func Log(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}

func Auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("k") != akey {
			http.Error(w, "forbidden", http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
}

func ackhandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Query().Get("h") == "" || r.URL.Query().Get("u") == "" {
		http.Error(w, "Invalid", http.StatusBadRequest)
		return
	}

	a := response.Ack{}

	a.User = r.URL.Query().Get("u")
	a.Key = r.URL.Query().Get("k")
	a.Host = r.URL.Query().Get("h")
	a.Service = r.URL.Query().Get("s")

	log.Printf("Ack{User: %s, Key: %s, Host: %s, Service: %s}\n", a.User, a.Key, a.Host, a.Service)

	if len(queue) >= 25 {
		queue = append(queue[:0], queue[1:]...)
	}

	queue = append(queue, a)
	log.Printf("In ackhandler queue(%p) = %d", &queue, cap(queue))

	fmt.Fprintf(w, "Success, %d commands queued.", len(queue))

}

func stats(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%d commands queued", len(queue))
}

func dequeue(w http.ResponseWriter, r *http.Request) {
	if queue != nil {
		j, _ := json.Marshal(queue)
		fmt.Fprintf(w, "%s", j)
		queue = nil
	}
	log.Printf("In dequeue queue(%p) = %d", &queue, cap(queue))

}

func main() {

	log.Printf("In main queue(%p) = %d", &queue, cap(queue))

	if len(akey) < 30 {
		log.Fatal("Must set NAGIOS_API_KEY to something pretty secure")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		http.Error(w, "nope", http.StatusForbidden)
	})
	http.HandleFunc("/stats", Auth(stats))
	http.HandleFunc("/dequeue", Auth(dequeue))
	http.HandleFunc("/ack", Auth(ackhandler))
	http.ListenAndServe(":8080", Log(http.DefaultServeMux))

}
