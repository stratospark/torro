package client

import (
	"fmt"
	"net"
	"time"
)

type BTConn struct{}

type Handler interface {
	StartListening(chan BTConn, error)
}

type BTService struct {
	Listener  *net.TCPListener
	Listening bool
	CloseCh   chan bool
	Port      int
}

func NewBTService(port int) *BTService {
	s := &BTService{
		Listening: false,
		CloseCh:   make(chan bool, 1),
		Port:      port,
	}
	return s
}

func (s *BTService) StartListening() (err error) {
	fmt.Println("Start listening")
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
				fmt.Println("Closing BitTorrent Service")
				s.Listener.Close()
				s.Listening = false
				return
			default:
			}

			fmt.Println("Before Accept")
			l.SetDeadline(time.Now().Add(time.Nanosecond))
			conn, err := l.AcceptTCP()
			fmt.Println("After Accept")
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println(conn)
			conn.Close()
		}
	}()

	return nil
}

func (s *BTService) StopListening() (err error) {
	fmt.Println("StopListening")
	s.CloseCh <- true
	return nil
}
