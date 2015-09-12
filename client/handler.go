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
	Conn           Connection
	State          BTState
	Hash           string
	PeerID         string
	HandshakeChan  chan bool
	MessageChan    chan bool
	WriteChan      chan *structure.BasicMessage
	DisconnectChan chan bool
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
	BTStateStartListening BTState = iota
	BTStateWaitingForHandshake
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
	AddChan           chan *BTConn
	LeaveChan         chan *BTConn
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
		AddChan:           make(chan *BTConn, 1),
		LeaveChan:         make(chan *BTConn, 1),
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
			//			go s.handleMessages()
			btc.handleConnection(s)
			btc.State = BTStateStartListening
			btc.HandshakeChan <- true
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
		btc, err := s.ConnectionFetcher.Dial(addr)
		if err != nil {
			// TODO: Try more than once before giving up?
			continue
		}
		hs, _ := structure.NewHandshake(hash, s.PeerID)
		btc.Write(hs.Bytes())
		btc.Hash = string(hash)
		btc.State = BTStateWaitingForHandshake
		btc.handleConnection(s)
		btc.HandshakeChan <- true
	}
}

func (s *BTService) handleMessages() {
	for {
		select {
		case d := <-s.DisconnectChan:
			if d {
				for btc := range s.Peers {
					btc.Close()
				}
			}
			s.TermCh <- true
		case btc := <-s.AddChan:
			log.Println("[handleMessages] Adding Peer")
			s.Peers[btc] = BTStateWaitingForHandshake
		case btc := <-s.LeaveChan:
			log.Println("[handleMessages] Removing Peer")
			delete(s.Peers, btc)
		}
	}
}

func (btc *BTConn) handleConnection(s *BTService) {
	btc.HandshakeChan = make(chan bool, 1)
	btc.MessageChan = make(chan bool, 1)
	btc.DisconnectChan = make(chan bool, 1)
	btc.WriteChan = make(chan *structure.BasicMessage, 1)
	btc.PeerID = string(s.PeerID)

	go btc.readLoop(s.AddChan, s.LeaveChan)
	go btc.writeLoop(s.AddChan, s.LeaveChan)

	return
}

func (btc *BTConn) readLoop(addChan, leaveChan chan<- *BTConn) {
	for {
		select {
		case _ = <-btc.HandshakeChan:
			log.Printf("[readLoop] Got Handshake\n")
			peerHs, err := handleHandshake(btc)
			if err != nil {
				log.Printf("[readLoop] error: %q", err.Error())
				btc.Close()
				leaveChan <- btc
				continue
			}

			log.Println(btc.State)
			switch btc.State {
			case BTStateWaitingForHandshake:
				log.Printf("%q === %q?", btc.Hash, string(peerHs.Hash))
				if btc.Hash != string(peerHs.Hash) {
					// TODO: What if same connection is handling multiple hashes?
					log.Printf("[readLoop] hash mismatch\n")
					btc.Close()
					leaveChan <- btc
					continue
				}
			case BTStateStartListening:
				log.Printf("Writing byte %q\n", btc)
				respHs, err := structure.NewHandshake(peerHs.Hash, []byte(btc.PeerID))
				log.Println("[readLoop] respHS ", respHs)
				if err != nil {
					log.Printf("[readLoop] %q\n", err.Error())
					btc.Close()
					leaveChan <- btc
					continue
				}
				btc.Write(respHs.Bytes())
			default:
				log.Printf("[readLoop] BAD STATE: %d", btc.State)
				btc.Close()
				leaveChan <- btc
				continue
			}

			addChan <- btc
			btc.State = BTStateReadyForMessages
			btc.MessageChan <- true
		case _ = <-btc.MessageChan:
			log.Printf("[readLoop] Reading Message from: %q", btc)
			m, err := structure.ReadMessage(btc)
			if err != nil {
				log.Printf("[readLoop] error reading message: %s", err)
				btc.Close()
				leaveChan <- btc
				continue
			}

			switch m.(type) {
			case *structure.BitFieldMessage:
				log.Println("BIT FIELD MESSAGE")
				btc.WriteChan <- &structure.BasicMessage{Type: structure.MessageTypeInterested, Length: 1}
			default:
				log.Println(" OTHER MESSAGE")
			}
			log.Println("[readLoop] got Mesage: ", m)
			time.Sleep(time.Millisecond * 100) // TODO: get rid of this sleep
			btc.Close()
			leaveChan <- btc
		}
	}
}

func (btc *BTConn) writeLoop(addChan, leaveChan chan<- *BTConn) {
	for msg := range btc.WriteChan {
		//		switch msg.Type {
		//
		//		}
		btc.Write(msg.Bytes())
	}
	return
}

func handleHandshake(c *BTConn) (hs *structure.Handshake, err error) {
	// First connection, assume handshake messsage
	// Get the protocol name length
	hs, err = structure.ReadHandshake(c)
	if err != nil {
		log.Printf("[handleHandshake] error: %q", err.Error())
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
	log.Println("StoppedListening")
	return nil
}
