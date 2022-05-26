package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nsqio/go-nsq"
)

type ProducerConf struct {
	AppKey    string `json:"app_key"`
	AppPort   string `json:"app_port"`
	NsqServer string `json:"nsq_server"`
	NsqTopic  string `json:"nsq_topic"`
	Debug     bool   `json:"debug"`
}

type StreamPayload struct {
	Tsclient int64  `json:"tsclient"`
	Tsserver int64  `json:"tsserver"`
	Sessid   string `json:"sessid"`
	Value    string `json:"value"`
	Label    string `json:"label"`
}

var confProducer ProducerConf = ProducerConf{}
var nsqProducer *nsq.Producer = nil

func main() {
	// Read config
	narg := len(os.Args)
	if narg < 2 {
		log.Fatalln("Missing producer config file")
		os.Exit(1)
	}
	cfile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln("Error open config file")
		os.Exit(1)
	}
	byteValue, _ := ioutil.ReadAll(cfile)
	json.Unmarshal(byteValue, &confProducer)
	cfile.Close()

	// Print config log
	fmt.Println("Config:")
	fmt.Println("  AppPort: ", confProducer.AppPort)
	fmt.Println("  NsqServer:", confProducer.NsqServer)
	fmt.Println("  NsqTopic: ", confProducer.NsqTopic)
	fmt.Println("  Debug: ", confProducer.Debug)

	nsqConfig := nsq.NewConfig()
	nsqProducer, err = nsq.NewProducer(confProducer.NsqServer, nsqConfig)
	if err != nil {
		if nsqProducer != nil {
			nsqProducer.Stop()
		}
		log.Fatalln("Error message system")
		os.Exit(1)
	}

	// HTTP Server
	r := mux.NewRouter()
	r.Use(CORSHandler)
	r.HandleFunc("/api/"+confProducer.AppKey, func(w http.ResponseWriter, r *http.Request) {
		var data []StreamPayload

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			fmt.Println(err.Error())
			http.Error(w, "JSON parsing error", http.StatusBadRequest)
			return
		}

		payload, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
			return
		}

		if err = nsqProducer.Publish(confProducer.NsqTopic, payload); err != nil {
			log.Println(err)
			return
		}
	})

	if err := http.ListenAndServe(":9000", r); err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}
}

func CORSHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
