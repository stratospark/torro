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
Connection to abstract over TCP, UDP, or mock sockets
*/
type Connection interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
}

/*
BTConn contains the information necessary to maintain a P2P
connection to a Peer according to the BitTorrent protocol.
*/
type BTConn struct {
	Conn   Connection
	PeerID string
}

func (btc *BTConn) Read(b []byte) (n int, err error) {
	return btc.Conn.Read(b)
}

func (btc *BTConn) Write(b []byte) (n int, err error) {
	return btc.Conn.Write(b)
}

func (btc *BTConn) Close() {
	btc.Conn.Close()
}

/*
ConnectionFetcher is an interface that returns a Connection
*/
type ConnectionFetcher interface {
	Dial(addr string) (*BTConn, error)
}

/*
TCPConnectionFetcher implements ConnectionFetcher using net.DialTCP
*/
type TCPConnectionFetcher struct {
}

func (t *TCPConnectionFetcher) Dial(addr string) (*BTConn, error) {
	conn, err := net.DialTimeout("tcp", addr, time.Millisecond*100)
	if err != nil {
		return nil, err
	}
	return &BTConn{Conn: conn}, nil
}

/*
Handler specifies the interface of a locally running BitTorrent client
*/
type Handler interface {
	StartListening(chan BTConn, error)
	AddHash([]byte)
}

/*
BTService is a wrapper around a TCPListener, along with
other state information.
*/
type BTService struct {
	ConnectionFetcher ConnectionFetcher
	Listener          *net.TCPListener
	Listening         bool
	CloseCh           chan bool
	TermCh            chan bool
	Port              int
	Peers             map[*BTConn]string
	Hashes            map[string]bool
	PeerID            []byte
}

/*
NewBTService returns a closed BTService on a specified port.
*/
func NewBTService(port int, peerId []byte) *BTService {
	s := &BTService{
		ConnectionFetcher: &TCPConnectionFetcher{},
		Listening:         false,
		CloseCh:           make(chan bool, 1),
		TermCh:            make(chan bool, 1),
		Port:              port,
		Peers:             make(map[*BTConn]string),
		Hashes:            make(map[string]bool),
		PeerID:            peerId,
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

	hsChan := make(chan *BTConn, 1)
	msgChan := make(chan string, 1)
	disconnectChan := make(chan bool, 1)

	go func() {
		go s.handleMessages(hsChan, msgChan, disconnectChan)

		for {
			select {
			case <-s.CloseCh:
				log.Println("Closing BitTorrent Service")
				disconnectChan <- true
				s.Listener.Close()
				s.Listening = false
				return
			default:
			}

			l.SetDeadline(time.Now().Add(time.Millisecond))
			conn, err := l.AcceptTCP()
			btc := &BTConn{Conn: conn}
			if err != nil {
				if !strings.Contains(err.Error(), "i/o timeout") {
					log.Println(err)
				}
				continue
			}
			log.Println(btc)
			go handleConnection(btc, hsChan)
		}
	}()

	return nil
}

func (s *BTService) AddHash(h []byte) {
	s.Hashes[string(h)] = true
}

func (s *BTService) InitiateHandshakes(hash []byte, peers []structure.Peer) {
	for _, peer := range peers {
		addr := fmt.Sprintf("%q:%d", peer.IP, peer.Port)
		conn, err := s.ConnectionFetcher.Dial(addr)
		if err != nil {
			continue
		}
		hs, _ := structure.NewHandshake(hash, s.PeerID)
		conn.Write(hs.Bytes())

		s.Peers[conn] = "added"
	}
}

func (s *BTService) handleMessages(hsChan <-chan *BTConn, msgChan <-chan string, disconnectChan <-chan bool) {
	for {
		select {
		case d := <-disconnectChan:
			if d {
				for k := range s.Peers {
					k.Close()
				}
			}
			s.TermCh <- true
		case hs := <-hsChan:
			peerHs, err := handleHandshake(hs)
			if err != nil {
				hs.Close()
				delete(s.Peers, hs)
				continue
			}

			// TODO: check if info has is the same
			s.Peers[hs] = "added"
			time.Sleep(time.Millisecond * 100)
			log.Printf("Writing byte %q\n", hs)
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

func handleConnection(c *BTConn, hsChan chan<- *BTConn) {
	log.Println("Handle Connection")

	hsChan <- c

	return
}

func handleHandshake(c *BTConn) (hs *structure.Handshake, err error) {
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
	_ = <-s.TermCh
	return nil
}
