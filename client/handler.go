package client

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

/*
BTConn contains the information necessary to maintain a P2P
connection to a Peer according to the BitTorrent protocol.
*/
type BTConn struct{}

type Handler interface {
	StartListening(chan BTConn, error)
}

/*
BTService is a wrapper around a TCPListener, along with
other state information.
*/
type BTService struct {
	Listener  *net.TCPListener
	Listening bool
	CloseCh   chan bool
	Port      int
}

/*
NewBTService returns a closed BTService on a specified port.
*/
func NewBTService(port int) *BTService {
	s := &BTService{
		Listening: false,
		CloseCh:   make(chan bool, 1),
		Port:      port,
	}
	return s
}

/*
StartListening starts a TCP listening service on a goroutine.
*/
func (s *BTService) StartListening() (err error) {
	log.Println("Start listening")
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", s.Port))
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	s.Listener = l
	s.Listening = true

	go func() {
		for {
			select {
			case <-s.CloseCh:
				log.Println("Closing BitTorrent Service")
				s.Listener.Close()
				s.Listening = false
				return
			default:
			}

			l.SetDeadline(time.Now().Add(time.Nanosecond))
			conn, err := l.AcceptTCP()
			if err != nil {
				if !strings.Contains(err.Error(), "i/o timeout") {
					log.Println(err)
				}
				continue
			}

			log.Println(conn)
			conn.Close()
		}
	}()

	return nil
}

/*
StopListening stops the TCP listener by sending to its Close channel.
*/
func (s *BTService) StopListening() (err error) {
	// TODO: Check that listener is actually on
	fmt.Println("StopListening")
	s.CloseCh <- true
	return nil
}
