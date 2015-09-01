package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

/*
BTConn contains the information necessary to maintain a P2P
connection to a Peer according to the BitTorrent protocol.
*/
type BTConn struct {
}

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

	hsChan := make(chan net.Conn, 1)
	msgChan := make(chan string, 1)

	go func() {
		go handleMessages(hsChan, msgChan)

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
			} else {
				log.Println(conn)
				go handleConnection(conn, hsChan)
			}

		}
	}()

	return nil
}

type Handshake struct {
	Length            byte
	Name              string
	ReservedExtension []byte
	Hash              []byte
	PeerID            []byte
}

func (h *Handshake) String() string {
	return fmt.Sprintf("pstrlen: %d, name: %s, reserved extension: %x , hash: %x , peer id: %s", h.Length, h.Name, h.ReservedExtension, h.Hash, h.PeerID)
}

func handleMessages(hsChan <-chan net.Conn, msgChan <-chan string) {
	peers := make(map[net.Conn]string)

	for {
		select {
		case hs := <-hsChan:
			handleHandshake(hs)
			peers[hs] = "added"
			log.Printf("Writing byte\n")
			hs.Write([]byte("pong"))
			hs.Close()
		default:

		}
	}
}

func handleConnection(c net.Conn, hsChan chan<- net.Conn) {
	log.Println("Handle Connection")

	hsChan <- c

	//	handleHandshake(c)

	//	log.Printf("Writing byte\n")
	//	c.Write([]byte("pong"))

	return
}

func handleHandshake(c net.Conn) {
	// First connection, assume handshake messsage
	// Get the protocol name length
	buf := make([]byte, 1)
	log.Println("Waiting to readfull")
	_, err := io.ReadFull(c, buf)
	if err != nil {
		log.Println("[HandleConnection] Error: ", err)
		return
	}
	pstrLen := int(buf[0])

	// Get the rest of the handshake message
	buf = make([]byte, pstrLen+48)
	_, err = io.ReadFull(c, buf)
	if err != nil {
		log.Println("[HandleConnection] Error: ", err)
		return
	}

	// Parse fields out of the message
	handshake := &Handshake{
		Length:            byte(pstrLen),
		Name:              string(buf[0:pstrLen]),
		ReservedExtension: buf[pstrLen : pstrLen+8],
		Hash:              buf[pstrLen+8 : pstrLen+8+20],
		PeerID:            buf[pstrLen+8+20 : pstrLen+8+20+20],
	}

	log.Printf("[HandleConnection] Handshake: %q", buf)
	log.Printf("%q", handshake)

	return
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
