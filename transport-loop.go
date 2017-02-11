// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"errors"
	"github.com/cooper/quiki/transport"
	"github.com/cooper/quiki/wikiclient"
	"log"
)

var tr transport.Transport

func initTransport() (err error) {

	// setup the transport
	tr, err = transport.New()
	if err != nil {
		return
	}

	// connect
	if err = tr.Connect(); err != nil {
		err = errors.New("can't connect to transport: " + err.Error())
		return
	}

	log.Println("connected to wikifier")
	tr.WriteMessage(wikiclient.NewMessage("wiki", map[string]interface{}{
		"name":     "notroll",
		"password": "hi",
	}))

	// start the loop
	go transportLoop()
	return
}

func transportLoop() {
	for {
		if tr == nil {
			panic("no transport?")
		}
		select {
		case err := <-tr.Errors():
			log.Println(err)
		case msg := <-tr.ReadMessages():
			log.Println(msg)
		}
	}
}
