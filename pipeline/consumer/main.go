package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
)

type ConsumerConf struct {
	NsqTopic    string `json:"nsq_topic"`
	NsqChannel  string `json:"nsq_channel"`
	NsqHostPort string `json:"nsq_host_port"`
	// DbConnString    string `json:"db_conn_string"`
	// DbTable         string `json:"db_table"`
	// MaxOpenConns    int    `json:"max_open_conns"`
	// MaxIdleConns    int    `json:"max_idle_conns"`
	// ConnMaxLifetime int    `json:"conn_max_lifetime"`
	Debug bool `json:"debug"`
}

type StreamPayload struct {
	Tsclient int64  `json:"tsclient"`
	Tsserver int64  `json:"tsserver"`
	Sessid   string `json:"sessid"`
	Value    string `json:"value"`
	Label    string `json:"label"`
}

var confConsumer ConsumerConf = ConsumerConf{}

type MyHandler struct{}

func (h *MyHandler) HandleMessage(message *nsq.Message) error {
	log.Printf("Got a message: %s", string(message.Body))
	return nil
}

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
	json.Unmarshal(byteValue, &confConsumer)
	cfile.Close()

	nsqConfig := nsq.NewConfig()

	// Maximum number of times this consumer will attempt to process a message before giving up
	nsqConfig.MaxAttempts = 10

	// Maximum number of messages to allow in flight
	nsqConfig.MaxInFlight = 5

	// Maximum duration when re-queueing
	nsqConfig.MaxRequeueDelay = time.Second * 900
	nsqConfig.DefaultRequeueDelay = time.Second * 0

	// Creating the consumer
	consumer, err := nsq.NewConsumer(confConsumer.NsqTopic, confConsumer.NsqChannel, nsqConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	// register our message handler with the consumer
	consumer.AddHandler(&MyHandler{})

	// connect to NSQ and start receiving messages
	//err = consumer.ConnectToNSQD("nsqd:4150")
	err = consumer.ConnectToNSQLookupd("127.0.0.1:4161")
	if err != nil {
		log.Fatal(err)
		return
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	// disconnect
	consumer.Stop()
}
