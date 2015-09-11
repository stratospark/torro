package client

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stratospark/torro/structure"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

type MockConnection struct {
	SendHandshakeChan chan *structure.Handshake
	SendMessageChan   chan structure.Message
	RestOfMessageChan chan []byte
	ReceiveBytesChan  chan []byte
}

func (c *MockConnection) SendHandshake(hs *structure.Handshake) {
	c.SendHandshakeChan <- hs
}

func (c *MockConnection) SendMessage(m structure.Message) {
	c.SendMessageChan <- m
}

func (c *MockConnection) Read(b []byte) (n int, err error) {
	readBytes := func(read []byte) {
		log.Printf("MockConnection Read: %q, len(read): %d, len(b): %d\n", read, len(read), len(b))
		for i := 0; i < len(b); i++ {
			b[i] = read[i]
		}
		if len(b) < len(read) {
			c.RestOfMessageChan <- read[len(b):]
		}
	}
	select {
	case read := <-c.RestOfMessageChan:
		readBytes(read)
	case hs := <-c.SendHandshakeChan:
		read := hs.Bytes()
		readBytes(read)
	case m := <-c.SendMessageChan:
		read := m.Bytes()
		readBytes(read)
	}

	return len(b), err
}

func (c *MockConnection) Write(b []byte) (n int, err error) {
	log.Printf("MockConnection Write: %q\n", b)
	c.ReceiveBytesChan <- b
	return len(b), nil
}

func (c *MockConnection) Close() error {
	return nil
}

type MockConnectionFetcher struct {
	Conns map[string]*BTConn
}

func NewMockConnectionFetcher() *MockConnectionFetcher {
	return &MockConnectionFetcher{
		Conns: make(map[string]*BTConn),
	}
}

func (t *MockConnectionFetcher) Dial(addr string) (*BTConn, error) {
	conn := &MockConnection{
		SendHandshakeChan: make(chan *structure.Handshake, 1),
		SendMessageChan:   make(chan structure.Message, 1),
		RestOfMessageChan: make(chan []byte, 1),
		ReceiveBytesChan:  make(chan []byte, 1),
	}
	btc := &BTConn{Conn: conn}
	t.Conns[addr] = btc

	return btc, nil
}

func TestHandler(t *testing.T) {
	port := 55555
	peerIdRemote := "-TR2840-nj5ovtREMOTE"
	peerIdClient := "-TR2840-nj5ovtCLIENT"
	hash := []byte("\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9")

	Convey("Listens to incoming connections on a given port", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		s.StartListening()

		time.Sleep(time.Millisecond)
		So(s.Listener, ShouldNotBeNil)
		So(s.Listening, ShouldBeTrue)

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})

	Convey("Accepts a handshake and adds to the connection list", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		s.AddHash(hash)
		s.StartListening()

		time.Sleep(time.Millisecond)
		So(s.Listener, ShouldNotBeNil)
		So(s.Listening, ShouldBeTrue)

		addr, _ := net.ResolveTCPAddr("tcp", "localhost:55555")
		conn, err := net.DialTCP("tcp", nil, addr)
		So(err, ShouldBeNil)

		handshake := "\x13\x42\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		conn.Write([]byte(handshake))
		time.Sleep(time.Millisecond)
		respHandshake, err := structure.ReadHandshake(conn)
		So(err, ShouldBeNil)
		So(respHandshake, ShouldNotBeNil)
		So(len(s.Peers), ShouldEqual, 1)

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})

	Convey("Rejects a malformed handshake request", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		s.StartListening()

		time.Sleep(time.Millisecond * 50)
		So(s.Listener, ShouldNotBeNil)
		So(s.Listening, ShouldBeTrue)

		addr, _ := net.ResolveTCPAddr("tcp", "localhost:55555")
		conn, err := net.DialTCP("tcp", nil, addr)
		So(err, ShouldBeNil)

		handshake := "\x13\x43\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		conn.Write([]byte(handshake))
		time.Sleep(time.Millisecond * 50)
		So(len(s.Peers), ShouldEqual, 0)
		buf := make([]byte, 4)
		_, err = io.ReadFull(conn, buf)
		So(err, ShouldNotBeNil)

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})

	Convey("Sends out handshake request to every IP in list", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		mc := NewMockConnectionFetcher()
		s.ConnectionFetcher = mc
		s.AddHash(hash)
		s.StartListening()

		time.Sleep(time.Millisecond * 50)

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
			c0.SendHandshake(hs)
		}

		time.Sleep(time.Millisecond * 50)
		So(len(s.Peers), ShouldEqual, len(peers))

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})

	Convey("Receives Bitfield message and sends Interested message", t, func() {
		s := NewBTService(port, []byte(peerIdRemote))
		mc := NewMockConnectionFetcher()
		s.ConnectionFetcher = mc
		s.AddHash(hash)
		s.StartListening()

		time.Sleep(time.Millisecond * 50)

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
			c0.SendHandshake(hs)

			bf := structure.BitFieldFromHexString("\xff\xff\xff\x01")
			msg := &structure.BitFieldMessage{BasicMessage: structure.BasicMessage{Type: structure.MessageTypeBitField, Length: 5, Payload: []byte("\xff\xff\xff\x01")}, BitField: bf}
			c0.SendMessage(msg)
		}

		time.Sleep(time.Millisecond * 50)

		_ = s.StopListening()
	})
}
