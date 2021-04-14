package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var value string

type Broker struct {
	clients        map[chan string]string
	newClients     chan chan string
	defunctClients chan chan string
	messages       chan string
}

// Start starts a new goroutine
func (b *Broker) Start() {
	go func() {
		for {
			select {

			case s := <-b.newClients:
				b.clients[s] = value
				log.Println("Added new client")

			case s := <-b.defunctClients:
				delete(b.clients, s)
				close(s)

				log.Println("Removed client")

			case msg := <-b.messages:
				// fmt.Println(msg)
				for c := range b.clients {
					c <- msg
				}
			}
		}
	}()
}

// This Broker method handles and HTTP request at the "/events/" URL.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	messageChan := make(chan string)

	b.newClients <- messageChan

	// Listen to the closing of the http connection via the CloseNotifier
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		b.defunctClients <- messageChan
		log.Println("HTTP connection just closed.")
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	for {
		msg, open := <-messageChan

		if !open {
			break
		}

		fmt.Fprintf(w, "data: %s\n\n", msg)

		f.Flush()
	}

	log.Println("Finished HTTP request at ", r.URL.Path)
}

// handler creates handler for /echo
func echoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/echo" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	value = r.URL.Query().Get("w")

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("WTF dude, error parsing your template.")

	}

	t.Execute(w, "")

	log.Println("Finished HTTP request at", r.URL.Path)
}

// sayHandler creates handler for /say
func sayHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/say" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	v := r.URL.Query().Get("w")
	value = v
}

func main() {
	b := &Broker{
		make(map[chan string]string),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
	}

	b.Start()

	http.Handle("/events/", b)

	go func() {
		for {
			b.messages <- value

			time.Sleep(time.Second * 1)
		}
	}()
	http.Handle("/echo", http.HandlerFunc(echoHandler))
	http.Handle("/say", http.HandlerFunc(sayHandler))

	fmt.Println("listen on port:8000")
	http.ListenAndServe(":8000", nil)
}
