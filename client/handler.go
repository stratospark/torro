package client

import (
	"fmt"
	"github.com/stratospark/torro/structure"
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
	AddHash([]byte)
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
	Peers     map[net.Conn]string
	Hashes    map[string]bool
	PeerID    []byte
}

/*
NewBTService returns a closed BTService on a specified port.
*/
func NewBTService(port int, peerId []byte) *BTService {
	s := &BTService{
		Listening: false,
		CloseCh:   make(chan bool, 1),
		Port:      port,
		Peers:     make(map[net.Conn]string),
		Hashes:    make(map[string]bool),
		PeerID:    peerId,
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
		go s.handleMessages(hsChan, msgChan)

		for {
			select {
			case <-s.CloseCh:
				log.Println("Closing BitTorrent Service")
				s.Listener.Close()
				s.Listening = false
				return
			default:
			}

			l.SetDeadline(time.Now().Add(time.Millisecond))
			conn, err := l.AcceptTCP()
			if err != nil {
				if !strings.Contains(err.Error(), "i/o timeout") {
					log.Println(err)
				}
				continue
			}
			log.Println(conn)
			go handleConnection(conn, hsChan)
		}
	}()

	return nil
}

func (s *BTService) AddHash(h []byte) {
	s.Hashes[string(h)] = true
}

func (s *BTService) handleMessages(hsChan <-chan net.Conn, msgChan <-chan string) {
	//	peers := make(map[net.Conn]string)

	for {
		select {
		case hs := <-hsChan:
			peerHs, err := handleHandshake(hs)
			if err != nil {
				time.Sleep(time.Millisecond * 100)
				hs.Close()
			}

			// TODO: check if info has is the same
			s.Peers[hs] = "added"
			time.Sleep(time.Millisecond * 100)
			log.Printf("Writing byte\n")
			respHs, err := structure.NewHandshake(peerHs.Hash, s.PeerID)
			log.Println("[handleMessages] respHS ", respHs)
			if err != nil {
				log.Printf("[handleMessages] %q\n", err.Error())
				hs.Close()
				delete(s.Peers, hs)
				continue
			}
			hs.Write(respHs.Bytes())
			time.Sleep(time.Second)
			hs.Close()
			delete(s.Peers, hs)
		}
	}
}

func handleConnection(c net.Conn, hsChan chan<- net.Conn) {
	log.Println("Handle Connection")

	hsChan <- c

	return
}

func handleHandshake(c net.Conn) (hs *structure.Handshake, err error) {
	// First connection, assume handshake messsage
	// Get the protocol name length
	hs, err = structure.ReadHandshake(c)
	if err != nil {
		return nil, err
	}

	//	log.Printf("[HandleConnection] Handshake: %q", buf)
	log.Printf("[handleHandshake] %q", hs)

	return hs, nil
}

/*
StopListening stops the TCP listener by sending to its Close channel.
*/
func (s *BTService) StopListening() (err error) {
	// TODO: Check that listener is actually on
	log.Println("StopListening")
	s.CloseCh <- true
	return nil
}
