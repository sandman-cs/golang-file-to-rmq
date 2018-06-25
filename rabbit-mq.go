package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sandman-cs/core"
	"github.com/streadway/amqp"
)

type chanToRabbit struct {
	payload string
	route   string
}

var (
	dstRabbitConn       *amqp.Connection
	dstRabbitCloseError chan *amqp.Error
	recCount            int
)

// Try to connect to the RabbitMQ server as
// long as it takes to establish a connection
//
func connectToRabbitMQ(uri string) *amqp.Connection {
	for {
		conn, err := amqp.Dial(uri)

		if err == nil {
			return conn
		}

		log.Println(err)
		log.Printf("Trying to reconnect to RabbitMQ at %s\n", uri)
		time.Sleep(500 * time.Millisecond)
	}
}

// re-establish the connection to RabbitMQ in case
// the connection has died
//
func dstRabbitConnector(uri string) {
	var dstRabbitErr *amqp.Error

	for {
		dstRabbitErr = <-dstRabbitCloseError
		if dstRabbitErr != nil {
			core.SendMessage(fmt.Sprintln("Connecting to Destination RabbitMQ: ", uri))
			dstRabbitConn = connectToRabbitMQ(uri)
			dstRabbitCloseError = make(chan *amqp.Error)
			dstRabbitConn.NotifyClose(dstRabbitCloseError)
			core.SendMessage("Connection to dst established...")
		}
	}

}

func pubToRabbit(payload string, route string) {

	var msg chanToRabbit
	msg.payload = payload
	msg.route = route

	messages <- msg
	return
}

func chanPubToRabbit(conn *amqp.Connection) {

	core.SendMessage("Opening Publish Channel from \"chanPubToRabbit\"...")
	ch, err := conn.Channel()
	core.SyslogCheckFuncError(err, "Error Opening Publish Channel")
	if err == nil {
		defer ch.Close()
		for {
			msg := <-messages
			body := msg.payload
			err = ch.Publish(
				conf.DstBrokerExchange, // exchange
				msg.route,              // routing key
				false,                  // mandatory
				false,                  // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				})
			core.SyslogCheckFuncError(err, "Publish Error")
			if err != nil {
				messages <- msg
				pubError++
				break
			} else {
				pubSuccess++
			}
		}
	} else {
		core.SendMessage("Connection Closed...")
	}
}
