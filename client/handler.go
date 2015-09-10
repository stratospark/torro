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

type BTState int

const (
	BTStateWaitingForHandshake BTState = iota
	BTStateReadyForMessages
)

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
	HsChan            chan *BTConn
	MsgChan           chan *BTConn
	DisconnectChan    chan bool
	Port              int
	Peers             map[*BTConn]BTState
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
		HsChan:            make(chan *BTConn, 1),
		MsgChan:           make(chan *BTConn, 1),
		DisconnectChan:    make(chan bool, 1),
		Port:              port,
		Peers:             make(map[*BTConn]BTState),
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

	go func() {
		go s.handleMessages()

		for {
			select {
			case <-s.CloseCh:
				log.Println("Closing BitTorrent Service")
				s.DisconnectChan <- true
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
			go handleConnection(btc, s.HsChan)
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
			// TODO: Try more than once before giving up?
			continue
		}
		hs, _ := structure.NewHandshake(hash, s.PeerID)
		conn.Write(hs.Bytes())

		s.Peers[conn] = BTStateWaitingForHandshake

		//		s.MsgChan <- conn

	}
}

func (s *BTService) handleMessages() {
	for {
		select {
		case d := <-s.DisconnectChan:
			if d {
				for k := range s.Peers {
					k.Close()
				}
			}
			s.TermCh <- true
		case conn := <-s.HsChan:
			peerHs, err := handleHandshake(conn)
			if err != nil {
				conn.Close()
				delete(s.Peers, conn)
				continue
			}

			// TODO: check if info has is the same
			log.Printf("Writing byte %q\n", conn)
			respHs, err := structure.NewHandshake(peerHs.Hash, s.PeerID)
			log.Println("[handleMessages] respHS ", respHs)
			if err != nil {
				log.Printf("[handleMessages] %q\n", err.Error())
				conn.Close()
				continue
			}
			s.Peers[conn] = BTStateReadyForMessages
			conn.Write(respHs.Bytes())
			s.MsgChan <- conn
		case conn := <-s.MsgChan:
			log.Printf("Reading Message from: %q", conn)
			time.Sleep(time.Millisecond * 100) // TODO: get rid of this sleep
			conn.Close()
			delete(s.Peers, conn)
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
