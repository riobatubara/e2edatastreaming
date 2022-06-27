package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"

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
var producer *nsq.Producer = nil

func main() {
	// Read config
	narg := len(os.Args)
	if narg < 2 {
		log.Fatalln("ERR::MISSING_ARGS")
		os.Exit(1)
	}

	cFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln("ERR::OPEN_CONFIG_FILE")
		os.Exit(1)
	}

	byteValue, _ := ioutil.ReadAll(cFile)
	json.Unmarshal(byteValue, &confProducer)
	cFile.Close()

	log.Println("Configuration loaded")
	log.Println("Connected to nsq server")

	config := nsq.NewConfig()
	producer, err = nsq.NewProducer(confProducer.NsqServer, config)
	if err != nil {
		if producer != nil {
			log.Println("Disconnect from nsq server")
			producer.Stop()
		}
		log.Fatalln("ERR::CONNECTING_NSQD")
		os.Exit(1)
	}

	// HTTP Server
	r := mux.NewRouter()
	r.Use(CORSHandler)
	r.HandleFunc("/api/"+confProducer.AppKey, func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

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

		if err = producer.Publish(confProducer.NsqTopic, payload); err != nil {
			log.Println(err)
			return
		}

		log.Printf("%v", data)
		log.Printf("alloc = %v MiB  totalAlloc = %v MiB  sys = %v MiB  numGC = %v MiB", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
		runtime.GC()
	})

	if err := http.ListenAndServe(":"+confProducer.AppPort, r); err != nil {
		log.Fatalf(err.Error())
		os.Exit(1)
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
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
