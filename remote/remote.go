package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:3500", "http service address")

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 20 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func main() {
	// flag.Parse()
	// log.SetFlags(0)

	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	done := make(chan struct{})
	doneR := make(chan struct{})

	for {

		log.Printf("connecting to %s", u.String())

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Println("dial:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		// defer c.Close()

		//write
		go func() {
			ticker := time.NewTicker(time.Second)
			tickerP := time.NewTicker(pingPeriod)

			defer ticker.Stop()
			defer tickerP.Stop()
			defer func() {
				log.Println("write closed")
			}()

			for {
				select {
				case <-done:
					return
				case t := <-ticker.C:
					c.SetWriteDeadline(time.Now().Add(writeWait))
					err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
					if err != nil {
						log.Println("write:", err)
						return
					}
				case <-tickerP.C:
					log.Println("sending ping")
					c.SetWriteDeadline(time.Now().Add(writeWait))
					if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}

				}
			}
		}()

		//read
		func() {

			c.SetReadLimit(maxMessageSize)
			c.SetReadDeadline(time.Now().Add(pongWait))
			c.SetPongHandler(func(string) error {
				c.SetReadDeadline(time.Now().Add(pongWait))
				fmt.Println("got a pong from client")
				return nil
			})
			// defer close(done)
			defer func() {
				log.Println("read closed")
			}()
			for {
				select {
				case <-doneR:
					return

				default:
					_, message, err := c.ReadMessage()
					if err != nil {
						log.Println("read:", err)
						done <- struct{}{}
						return
					}
					log.Printf("recv: %s", message)
				}
			}
		}()
	}
}
