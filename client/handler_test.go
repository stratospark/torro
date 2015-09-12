package client

import (
	"bytes"
	"errors"
	"github.com/oleiade/lane"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stratospark/torro/structure"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

/*
MockConnection facilitates testing a socket connection by
having a SendMessageChan that sends a given message to a peer,
and a ReceiveBytesChan where a test can assert a valid response.
*/
type MockConnection struct {
	ReadQueue         *lane.Queue
	SendMessageChan   chan structure.Message
	RestOfMessageChan chan bool
	ReceiveBytesChan  chan []byte
}

/*
SendMessage puts a message on the channel to be sent to the peer.
*/
func (c *MockConnection) SendMessage(m structure.Message) {
	c.SendMessageChan <- m
}

/*
Read fills the []byte b with a portion of the next sent message.
If only a portion of the message was sent, the head of the queue
is modified to be the remainder of that message.
*/
func (c *MockConnection) Read(b []byte) (n int, err error) {
	log.Println("Trying to read, len(b): ", len(b))
	readBytes := func() {
		log.Printf("ReadBytes, len(queue): %d", c.ReadQueue.Size())
		h := c.ReadQueue.Dequeue()
		read, _ := h.([]byte)
		log.Printf("MockConnection Read: %q, len(read): %d, len(b): %d\n", read, len(read), len(b))
		for i := 0; i < len(b); i++ {
			b[i] = read[i]
		}
		log.Printf("After REad")
		if len(b) < len(read) {
			c.ReadQueue.Prepend(read[len(b):])
			c.RestOfMessageChan <- true
		}
	}
	select {
	case <-c.RestOfMessageChan:
		readBytes()
	case m := <-c.SendMessageChan:
		log.Println("SendMessageChan")
		read := m.Bytes()
		c.ReadQueue.Enqueue(read)
		readBytes()
	}

	return len(b), err
}

/*
Write sends a []byte b to the receive channel where a test can
assert whether it is valid or not.
*/
func (c *MockConnection) Write(b []byte) (n int, err error) {
	log.Printf("MockConnection Write: %q\n", b)
	c.ReceiveBytesChan <- b
	return len(b), nil
}

/*
Close is a no-op stub request
*/
func (c *MockConnection) Close() error {
	return nil
}

/*
MockConnectionFetcher will return MockConnections rather than
TCPConnections to facilitate testing without the network.
*/
type MockConnectionFetcher struct {
	Conns map[string]*BTConn
}

func NewMockConnectionFetcher() *MockConnectionFetcher {
	return &MockConnectionFetcher{
		Conns: make(map[string]*BTConn),
	}
}

/*
Dial returns a new MockConnection and adds it to the set of Conns.
*/
func (t *MockConnectionFetcher) Dial(addr string) (*BTConn, error) {
	conn := &MockConnection{
		ReadQueue:         lane.NewQueue(),
		SendMessageChan:   make(chan structure.Message, 1),
		RestOfMessageChan: make(chan bool, 1),
		ReceiveBytesChan:  make(chan []byte, 1),
	}
	btc := &BTConn{Conn: conn}
	t.Conns[addr] = btc

	return btc, nil
}

var (
	port         int    = 55555
	peerIdRemote string = "-TR2840-nj5ovtREMOTE"
	peerIdClient string = "-TR2840-nj5ovtCLIENT"
	hash         []byte = []byte("\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9")
)

func TestListen(t *testing.T) {
	Convey("Listens to incoming connections on a given port", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		s.StartListening()

		So(s.Listener, ShouldNotBeNil)
		So(s.Listening, ShouldBeTrue)

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})

}

func TestAcceptHandshake(t *testing.T) {
	Convey("Accepts a handshake and adds to the connection list", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		s.AddHash(hash)
		s.StartListening()

		So(s.Listener, ShouldNotBeNil)
		So(s.Listening, ShouldBeTrue)

		addr, _ := net.ResolveTCPAddr("tcp", "localhost:55555")
		conn, err := net.DialTCP("tcp", nil, addr)
		So(err, ShouldBeNil)

		handshake := "\x13\x42\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		conn.Write([]byte(handshake))
		respHandshake, err := structure.ReadHandshake(conn)
		So(err, ShouldBeNil)
		So(respHandshake, ShouldNotBeNil)
		So(len(s.Peers), ShouldEqual, 1)

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})
}

func TestRejectHandshake(t *testing.T) {
	Convey("Rejects a malformed handshake request", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		s.StartListening()

		So(s.Listener, ShouldNotBeNil)
		So(s.Listening, ShouldBeTrue)

		addr, _ := net.ResolveTCPAddr("tcp", "localhost:55555")
		conn, err := net.DialTCP("tcp", nil, addr)
		So(err, ShouldBeNil)

		handshake := "\x13\x43\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		conn.Write([]byte(handshake))
		So(len(s.Peers), ShouldEqual, 0)
		buf := make([]byte, 4)
		_, err = io.ReadFull(conn, buf)
		So(err, ShouldNotBeNil)

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})
}

func TestInitiateHandshakes(t *testing.T) {
	Convey("Sends out handshake request to every IP in list", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		mc := NewMockConnectionFetcher()
		s.ConnectionFetcher = mc
		s.AddHash(hash)
		s.StartListening()

		// TODO: check that peer data is saved within service data structure
		peers := make([]structure.Peer, 2)
		peers[0] = structure.Peer{IP: net.IPv4(192, 168, 1, 1), Port: 55556}
		peers[1] = structure.Peer{IP: net.IPv4(192, 168, 1, 2), Port: 55557}
		s.InitiateHandshakes(hash, peers)

		for _, p := range peers {
			c0 := mc.Conns[p.AddrString()].Conn.(*MockConnection)
			hs, err := structure.ReadHandshake(bytes.NewReader(<-c0.ReceiveBytesChan))
			So(hs, ShouldNotBeNil)
			So(err, ShouldBeNil)
			hs, _ = structure.NewHandshake(hash, []byte(peerIdClient))
			c0.SendMessage(hs)
		}

		time.Sleep(time.Millisecond)
		So(len(s.Peers), ShouldEqual, len(peers))

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})
}

func ReadMessageOrTimeout(c *MockConnection) (*structure.BasicMessage, error) {
	select {
	case b := <-c.ReceiveBytesChan:
		m, err := structure.ReadMessage(bytes.NewReader(b))
		bm, _ := m.(*structure.BasicMessage)
		So(bm, ShouldNotBeNil)
		So(err, ShouldBeNil)
		return bm, err
	case <-time.After(time.Millisecond * 10):
		return nil, errors.New("Timeout")
	}
}

func TestConversation(t *testing.T) {
	Convey("Receives Bitfield message and sends Interested message", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		mc := NewMockConnectionFetcher()
		s.ConnectionFetcher = mc
		s.AddHash(hash)
		s.StartListening()

		// TODO: check that peer data is saved within service data structure
		peers := make([]structure.Peer, 1)
		peers[0] = structure.Peer{IP: net.IPv4(192, 168, 1, 1), Port: 55556}
		s.InitiateHandshakes(hash, peers)

		for _, p := range peers {
			c0 := mc.Conns[p.AddrString()].Conn.(*MockConnection)
			hs, err := structure.ReadHandshake(bytes.NewReader(<-c0.ReceiveBytesChan))
			So(hs, ShouldNotBeNil)
			So(err, ShouldBeNil)

			hs, _ = structure.NewHandshake(hash, []byte(peerIdClient))
			c0.SendMessage(hs)

			bf := structure.BitFieldFromHexString("\xff\xff\xff\x01")
			msg := &structure.BitFieldMessage{BasicMessage: structure.BasicMessage{Type: structure.MessageTypeBitField, Length: 5, Payload: []byte("\xff\xff\xff\x01")}, BitField: bf}
			c0.SendMessage(msg)

			m, err := ReadMessageOrTimeout(c0)
			So(err, ShouldBeNil)
			So(m.Type, ShouldEqual, structure.MessageTypeInterested)
		}

		_ = s.StopListening()
	})
}
