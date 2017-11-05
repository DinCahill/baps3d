package list

// This file defines Controller, a wrapper for List that exposes a channel interface.
// It is used to lift a List into a goroutine communicating with players and clients.
// For the protocol used by the Controller, see 'messages.go'.

import (
	"fmt"
	"reflect"
)

// Controller wraps a List in a channel-based interface.
type Controller struct {
	// list is the internal list managed by the Controller.
	list *List

	// clients is the list of Controller-facing client channel pairs.
	// Each client that subscribes gets a Client struct with the other sides.
	clients map[coclient]struct{}
}

// Client is the type of external Controller client handles.
type Client struct {
	// Tx is the channel through which the Client can send requests to the Controller.
	Tx chan<- Request

	// Rx is the channel on which the Controller sends status update messages.
	Rx <-chan Response
}

// coclient is the type of internal client handles.
type coclient struct {
	// tx is the status update send channel.
	tx chan<- Response

	// rx is the request receiver channel.
	rx <-chan Request
}

type Response struct {
	Body Responder
}

type Responder interface{}

// makeClient creates a new client and coclient pair.
func makeClient() (Client, coclient) {
	rq := make(chan Request)
	rs := make(chan Response)
	client := Client{Tx: rq, Rx: rs}
	coclient := coclient{tx: rs, rx: rq}
	return client, coclient
}

// NewController constructs a new Controller for a given List.
func NewController(l *List) (*Controller, *Client) {
	client, co := makeClient()

	coclients := make(map[coclient]struct{})
	coclients[co] = struct{}{}

	controller := Controller{
		list:    l,
		clients: coclients,
	}

	return &controller, &client
}

// Run runs this Controller's event loop.
func (c *Controller) Run() {
	cases := make([]reflect.SelectCase, len(c.clients))
	i := 0
	for cl := range c.clients {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(cl.rx)}
		i++
	}

	for {
		// TODO(@MattWindsor91): recalculate client cases when forking

		_, value, ok := reflect.Select(cases)
		if !ok {
			break
		}
		// TODO(@MattWindsor91): properly handle if this isn't a Request
		rq, ok := value.Interface().(Request)
		if !ok {
			fmt.Println("FIXME: got bad request")
		}
		
		c.handleRequest(rq)
	}

	c.hangupClients()
}

// hangupClients hangs up every connected client.
func (c *Controller) hangupClients() {
	for cl := range c.clients {
		c.hangupClient(cl)
	}
}

// hangupClient closes a client's channels and removes it from the client list.
func (c *Controller) hangupClient(cl coclient) {
	close(cl.tx)
	delete(c.clients, cl)
}

// handleRequest handles a request.
func (c *Controller) handleRequest(rq Request) {
	// TODO(@MattWindsor91): send unicast responses back
	switch body := rq.Body.(type) {
	case SetAutoModeRequest:
		c.list.SetAutoMode(body.AutoMode)
	}
}
