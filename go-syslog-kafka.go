package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mcuadros/go-syslog"

	"log"
)

// Program parameters
var adminIface string
var tcpIface string
var udpIface string
var kBrokers string
var kBatch int
var kTopic string
var kBufferTime int
var kBufferBytes int
var pEnabled bool
var cCapacity int

// Diagnostic data
var startTime time.Time

// Types
const (
	adminHost        = "localhost:8080"
	connTcpHost      = "localhost:514"
	connUdpHost      = "localhost:514"
	connType         = "tcp"
	kafkaBatch       = 10
	kafkaBrokers     = "localhost:9092"
	kafkaTopic       = "logs"
	kafkaBufferTime  = 1000
	kafkaBufferBytes = 512 * 1024
	parseEnabled     = true
	chanCapacity     = 0
)

func init() {
	flag.StringVar(&adminIface, "admin", adminHost, "Admin interface")
	flag.StringVar(&tcpIface, "tcp", connTcpHost, "TCP bind interface")
	flag.StringVar(&udpIface, "udp", connUdpHost, "UDP interface")
	flag.StringVar(&kBrokers, "broker", kafkaBrokers, "comma-delimited kafka brokers")
	flag.StringVar(&kTopic, "topic", kafkaTopic, "kafka topic")
	flag.IntVar(&kBatch, "batch", kafkaBatch, "Kafka batch size")
	flag.IntVar(&kBufferTime, "maxbuff", kafkaBufferTime, "Kafka client buffer max time (ms)")
	flag.IntVar(&kBufferBytes, "maxbytes", kafkaBufferBytes, "Kafka client buffer max bytes")
	flag.BoolVar(&pEnabled, "parse", parseEnabled, "enable syslog header parsing")
	flag.IntVar(&cCapacity, "chancap", chanCapacity, "channel buffering capacity")
}

// isPretty returns whether the HTTP response body should be pretty-printed.
func isPretty(req *http.Request) (bool, error) {
	err := req.ParseForm()
	if err != nil {
		return false, err
	}
	if _, ok := req.Form["pretty"]; ok {
		return true, nil
	}
	return false, nil
}

func main() {
	flag.Parse()

	startTime = time.Now()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("unable to determine hostname -- aborting")
		os.Exit(1)
	}
	log.Printf("syslog server starting on %s, PID %d", hostname, os.Getpid())
	log.Printf("machine has %d cores", runtime.NumCPU())

	// Log config
	log.Printf("kafka brokers: %s", kBrokers)
	log.Printf("kafka topic: %s", kTopic)
	log.Printf("kafka batch size: %d", kBatch)
	log.Printf("kafka buffer time: %dms", kBufferTime)
	log.Printf("kafka buffer bytes: %d", kBufferBytes)
	log.Printf("parsing enabled: %t", pEnabled)
	log.Printf("channel buffering capacity: %d", cCapacity)

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC5424)
	server.SetHandler(handler)
	server.ListenUDP(connUdpHost)
	server.Boot()

	// Connect to Kafka
	_, err = NewKafkaProducer(channel, strings.Split(kBrokers, ","), kTopic, kBufferTime, kBufferBytes)
	if err != nil {
		fmt.Println("Failed to create Kafka producer", err.Error())
		os.Exit(1)
	}
	log.Printf("connected to kafka at %s", kBrokers)

	server.Wait()
}
