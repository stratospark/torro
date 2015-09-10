package client

import (
	"github.com/oleiade/lane"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stratospark/torro/structure"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

type MockConnection struct {
	Stack *lane.Stack
}

func (c *MockConnection) Read(b []byte) (n int, err error) {
	read := c.Stack.Pop().([]byte)
	log.Printf("MockConnection Read: %q\n", read)
	for i := 0; i < len(b); i++ {
		b[i] = read[i]
	}
	c.Stack.Push(read[len(b):])
	return len(b), err
}

func (c *MockConnection) Write(b []byte) (n int, err error) {
	log.Printf("MockConnection Write: %q\n", b)
	c.Stack.Push(b)
	log.Printf("%d", c.Stack.Size())
	return len(b), nil
}

func (c *MockConnection) Close() error {
	return nil
}

type MockConnectionFetcher struct{}

func (t *MockConnectionFetcher) Dial(addr string) (*BTConn, error) {
	conn := &MockConnection{Stack: lane.NewStack()}
	return &BTConn{Conn: conn}, nil
}

func TestHandler(t *testing.T) {
	port := 55555
	peerId := "-TR2840-nj5ovtkoz2ed8"
	hash := []byte("\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9")

	Convey("Listens to incoming connections on a given port", t, func() {
		s := NewBTService(port, []byte(peerId))
		s.StartListening()

		time.Sleep(time.Millisecond)
		So(s.Listener, ShouldNotBeNil)
		So(s.Listening, ShouldBeTrue)

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})

	Convey("Accepts a handshake and adds to the connection list", t, func() {
		s := NewBTService(port, []byte(peerId))
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
		s := NewBTService(port, []byte(peerId))
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
		t.Log("InitHandshake test")
		s := NewBTService(port, []byte(peerId))
		s.ConnectionFetcher = &MockConnectionFetcher{}
		s.AddHash(hash)
		s.StartListening()

		time.Sleep(time.Millisecond * 50)

		So(true, ShouldBeTrue)

		// TODO: check that peer data is saved within service data structure
		peers := make([]structure.Peer, 2)
		peers[0] = structure.Peer{IP: net.IPv4(192, 168, 1, 1), Port: 55556}
		peers[1] = structure.Peer{IP: net.IPv4(192, 168, 1, 2), Port: 55557}
		s.InitiateHandshakes(hash, peers)
		So(len(s.Peers), ShouldEqual, len(peers))

		_ = s.StopListening()
		So(s.Listening, ShouldBeFalse)
	})
}
