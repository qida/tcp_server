package tcp_server

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

const (
	REPLAY_HOST = "root.sunqida.cn:7005"
)

// Client holds info about connection
type Client struct {
	Id          int
	conn        net.Conn
	conn_replay net.Conn
	Server      *server
	incoming    chan string // Channel for incoming data from client
}

// TCP server
type server struct {
	clients                  []*Client
	address                  string        // Address to open connection: localhost:9999
	IsReplay                 bool          // Address to open connection: localhost:9999
	joins                    chan net.Conn // Channel for new connections
	onNewClientCallback      func(c *Client)
	onClientConnectionClosed func(c *Client, err error)
	onNewMessage             func(c *Client, message string)
}

// Read client data from channel
func (c *Client) listen() {
	reader := bufio.NewReader(c.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			c.conn.Close()
			c.Server.onClientConnectionClosed(c, err)
			return
		}
		c.Server.onNewMessage(c, message)
		c.incoming <- message
	}
}

//转发
func (c *Client) replay() {
	c.incoming = make(chan string)
	for {
		fmt.Println("replay start")
		msg := <-c.incoming
		if c.conn_replay == nil {
			var err error
			c.conn_replay, err = net.Dial("tcp", REPLAY_HOST)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			c.conn_replay.Write([]byte(msg))
			fmt.Printf("转发 %s\r\n", msg)
			fmt.Println("================")
		}
		// _, err := io.Copy(c.conn_replay, c.conn)
		// if err != nil {
		// 	fmt.Println(err.Error())
		// 	return
		// }
	}
}

func (c *Client) Send(message string) error {
	_, err := c.conn.Write([]byte(message + "\n"))
	return err
}

// Get conn
func (c *Client) GetConn() net.Conn {
	return c.conn
}

// Called right after server starts listening new client
func (s *server) OnNewClient(callback func(c *Client)) {
	s.onNewClientCallback = callback
}

// Called right after connection closed
func (s *server) OnClientConnectionClosed(callback func(c *Client, err error)) {
	s.onClientConnectionClosed = callback
}

// Called when Client receives new message
func (s *server) OnNewMessage(callback func(c *Client, message string)) {
	s.onNewMessage = callback
}

// Creates new Client instance and starts listening
func (s *server) newClient(conn net.Conn) {
	client := &Client{
		conn:   conn,
		Server: s,
	}
	go client.listen()
	go client.replay()
	s.onNewClientCallback(client)
}

// Listens new connections channel and creating new client
func (s *server) listenChannels() {
	for {
		select {
		case conn := <-s.joins:
			s.newClient(conn)
		}
	}
}

// Creates new tcp server instance
func New(address string, is_relay bool) *server {
	log.Println("Creating server with address", address)
	server := &server{
		address: address,
		joins:   make(chan net.Conn),
	}
	server.OnNewClient(func(c *Client) {})
	server.OnNewMessage(func(c *Client, message string) {})
	server.OnClientConnectionClosed(func(c *Client, err error) {})
	return server
}
