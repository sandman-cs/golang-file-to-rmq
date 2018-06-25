package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type tailFile struct {
	Path  string
	Route string
	Ext   string
}

// Configuration File Opjects
type configuration struct {
	SrcDir            []tailFile
	ServerName        string
	AppName           string
	AppVer            string
	ChannelSize       int
	ChannelCount      int
	LogLevel          string
	DstBroker         string
	DstBrokerUsr      string
	DstBrokerPwd      string
	DstBrokerExchange string
	DstBrokerVhost    string
}

var (
	conf       configuration
	messages   chan chanToRabbit
	pubSuccess int
	pubError   int
)

func init() {

	//Load Default Configuion Values

	conf.AppName = "Go - File Exporter"
	conf.AppVer = "0.1"
	conf.ServerName, _ = os.Hostname()
	conf.ChannelSize = 1024
	conf.ChannelCount = 1
	conf.LogLevel = "info"

	//Load Configuration Data
	dat, _ := ioutil.ReadFile("conf.json")
	err := json.Unmarshal(dat, &conf)
	if err != nil {
		log.Println("Failed to load config.json")
		os.Exit(0)
	}

	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	//Connect to Destination RMQ Broker and Exchange.......
	//Channel Setup
	fmt.Println("Creating Message Channel...")
	messages = make(chan chanToRabbit, conf.ChannelSize)

	// create the rabbitmq error channel
	fmt.Println("Creating RabbitMQ Close Error Channel...")
	dstRabbitCloseError = make(chan *amqp.Error)

	// run the callback in a separate thread
	fmt.Println("Creating RabbitMQ Callback Channel...")
	go dstRabbitConnector(fmt.Sprint("amqp://" + conf.DstBrokerUsr + ":" + conf.DstBrokerPwd + "@" + conf.DstBroker + conf.DstBrokerVhost))

	// establish the rabbitmq connection by sending
	// an error and thus calling the error callback
	fmt.Println("Sending close error to initiate connection...")
	dstRabbitCloseError <- amqp.ErrClosed

	for dstRabbitConn == nil {
		fmt.Println("Waiting for connection to rabbitmq...")
		time.Sleep(time.Second * 1)
	}

	for i := 0; i < conf.ChannelCount; i++ {
		go func() {
			for {
				chanPubToRabbit(dstRabbitConn)
				time.Sleep(time.Second * 5)
			}
		}()
	}

	//Keep this part, spawn all the cool new stuff...............................
	for index, element := range conf.SrcDir {
		// Create Channel and launch mail publish threads.......
		fmt.Println("Creating Channel #", index)
		fmt.Println("For: ", element.Path)
		go workLoop(element.Path, element.Route, element.Ext)

	}

}

//checkError function
func checkError(err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"app":    conf.AppName,
			"server": conf.ServerName,
		}).Error(fmt.Sprint(err))
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

//sendMessage to udp listener
func sendMessage(msg string) {

	log.WithFields(log.Fields{
		"app":    conf.AppName,
		"ver":    conf.AppVer,
		"server": conf.ServerName,
	}).Info(msg)
}

//sendMessage to udp listener
func sendDebugMessage(msg string) {

	log.WithFields(log.Fields{
		"app":    conf.AppName,
		"ver":    conf.AppVer,
		"server": conf.ServerName,
	}).Debug(msg)
}

//sendMessage to udp listener
func sendWarnMessage(msg string) {

	log.WithFields(log.Fields{
		"app":    conf.AppName,
		"ver":    conf.AppVer,
		"server": conf.ServerName,
	}).Warn(msg)
}
