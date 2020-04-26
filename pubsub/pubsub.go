package pubsub

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

const (
	PUBLISH     = "publish"
	UNSUBSCRIBE = "unsubscribe"
	SUBSCRIBE   = "subscribe"
)

// PubSub - struct
type PubSub struct {
	Clients       []Client
	Subscriptions []Subscription
}

// Client - struct
type Client struct {
	ID         string
	Connection *websocket.Conn
}

// Message - struct
type Message struct {
	Action  string          `json:"action"`
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

// Subscription - struct
type Subscription struct {
	Topic  string
	Client *Client
}

// AddClient - struct add new Client
func (ps *PubSub) AddClient(c Client) *PubSub {

	ps.Clients = append(ps.Clients, c)

	payload := []byte("Hello Client ID" + c.ID)
	c.Connection.WriteMessage(1, payload)
	return ps
}

// Publish - ps topic publish
func (ps *PubSub) Publish(topic string, message []byte, excludeClient *Client) {

	subscription := ps.GetSubscriptions(topic, nil)

	for _, sub := range subscription {
		sub.Client.Send(message)
	}
}

// Send - send message pub/sub
func (c *Client) Send(message []byte) error {
	return c.Connection.WriteMessage(1, message)
}

// Unsubscribe - handler unsubscribes
func (ps *PubSub) Unsubscribe(c *Client, topic string) *PubSub {
	// clientSubscription := ps.GetSubscriptions(topic, c)

	for i, sub := range ps.Subscriptions {
		if sub.Client.ID == c.ID && sub.Topic == topic {
			// found this subscription from client anda we do need remove it
			ps.Subscriptions = append(ps.Subscriptions[:i], ps.Subscriptions[i+1:]...)
		}
	}

	return ps
}

// RemoveClient - remove subscription and client pub/sub
func (ps *PubSub) RemoveClient(c Client) *PubSub {

	// first remove all subscriptions by this client
	for i, sub := range ps.Subscriptions {
		if c.ID == sub.Client.ID {
			ps.Subscriptions = append(ps.Subscriptions[:i], ps.Subscriptions[i+1:]...)
		}
	}
	for index, client := range ps.Clients {
		if c.ID == client.ID {
			ps.Clients = append(ps.Clients[:index], ps.Clients[index+1:]...)
		}
	}
	// remove client from client
	return ps
}

// Subscribe - function handler Subscribe
func (ps *PubSub) Subscribe(client *Client, topic string) *PubSub {

	clientSubs := ps.GetSubscriptions(topic, client)

	if len(clientSubs) > 0 {
		// client is subscribe
		return ps
	}

	newSubscription := Subscription{
		Topic:  topic,
		Client: client,
	}

	ps.Subscriptions = append(ps.Subscriptions, newSubscription)
	return ps
}

// GetSubscriptions - client
func (ps *PubSub) GetSubscriptions(topic string, client *Client) []Subscription {

	var subscriptionList []Subscription
	for _, subscription := range ps.Subscriptions {
		if client != nil {
			if subscription.Client.ID == client.ID && subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		} else {
			if subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		}
	}
	return subscriptionList
}

// HandleReceiveMessage - struct
func (ps *PubSub) HandleReceiveMessage(c Client, messageType int, message []byte) *PubSub {

	m := Message{}

	err := json.Unmarshal(message, &m)

	if err != nil {
		return ps
	}

	switch m.Action {
	case PUBLISH:
		ps.Publish(m.Topic, m.Message, &c)
		break
	case UNSUBSCRIBE:
		ps.Unsubscribe(&c, m.Topic)
		log.Println("new unsubscribes to topic", m.Topic, len(ps.Subscriptions), c.ID)
	case SUBSCRIBE:
		ps.Subscribe(&c, m.Topic)
		log.Println("new subscriber to topic", m.Topic, len(ps.Subscriptions), c.ID)
		break
	default:
		break
	}

	return ps
}
