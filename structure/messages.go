package structure

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

var (
	ErrNotBitTorrentProtocol error = errors.New("Not BitTorrentProtocol")
)

type Reader interface {
	Read(p []byte) (n int, err error)
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

func NewHandshake(hash, peerId []byte) (h *Handshake, err error) {
	return &Handshake{
		Length:            19,
		Name:              "BitTorrent protocol",
		ReservedExtension: []byte("\x00\x00\x00\x00\x00\x00\x00\x00"),
		Hash:              hash,
		PeerID:            peerId,
	}, nil
}

/*
ReadHandshake pulls the next message off of the Reader and
verifies that it conforms to the BitTorrent Protocol.
*/
func ReadHandshake(r Reader) (h *Handshake, err error) {
	buf := make([]byte, 1)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		log.Println("[ReadHandshake] ReadFull Error: ", err)
		return nil, err
	}
	pstrLen := int(buf[0])
	log.Println(pstrLen)

	// Get the rest of the handshake message
	buf = make([]byte, pstrLen+48)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		// Fewer bytes than expected?
		log.Printf("[ReadHandshake] More ReadFull Error: %s, %x, %d", err, buf, len(buf))
		return nil, err
	}

	name := string(buf[0:pstrLen])
	log.Printf(name)
	if name != "BitTorrent protocol" {
		log.Println("[ReadHandshake] Not BitTorrent protocol handshake")
		return nil, ErrNotBitTorrentProtocol
	}

	// Parse fields out of the message
	h = &Handshake{
		Length:            byte(pstrLen),
		Name:              string(buf[0:pstrLen]),
		ReservedExtension: buf[pstrLen : pstrLen+8],
		Hash:              buf[pstrLen+8 : pstrLen+8+20],
		PeerID:            buf[pstrLen+8+20 : pstrLen+8+20+20],
	}

	log.Println("[ReadHandshake]: ", h)

	return h, nil
}

/*
Bytes serializes the Handshake message to []byte.
*/
func (h *Handshake) Bytes() []byte {
	bs := make([]byte, 0)
	buf := bytes.NewBuffer(bs)
	buf.Write([]byte{h.Length})
	buf.Write([]byte(h.Name))
	buf.Write(h.ReservedExtension)
	buf.Write(h.Hash)
	buf.Write(h.PeerID)
	return buf.Bytes()
}

type MessageType byte

const (
	MessageTypeKeepAlive     MessageType = 255
	MessageTypeChoke         MessageType = 0
	MessageTypeUnchoke       MessageType = 1
	MessageTypeInterested    MessageType = 2
	MessageTypeNotInterested MessageType = 3
	MessageTypeHave          MessageType = 4
	MessageTypeBitField      MessageType = 5
	MessageTypeRequest       MessageType = 6
	MessageTypePiece         MessageType = 7
	MessageTypeCancel        MessageType = 8
	MessageTypePort          MessageType = 9
)

type Message interface {
	Bytes() []byte
}

type BasicMessage struct {
	Length  int
	Type    MessageType
	Payload []byte
}

type HaveMessage struct {
	BasicMessage
	PieceIndex int
}

type BitFieldMessage struct {
	BasicMessage
	BitField *BitField
}

type RequestMessage struct {
	BasicMessage
	PieceIndex  int
	BeginOffset int
	PieceLength int
}

type PieceMessage struct {
	BasicMessage
	PieceIndex  int
	BeginOffset int
	Block       []byte
}

type CancelMessage struct {
	BasicMessage
	PieceIndex  int
	BeginOffset int
	PieceLength int
}

type PortMessage struct {
	BasicMessage
	Port int
}

/*
Bytes converts a message into its []byte representation, useful
for serializing over the wire.
*/
func (m BasicMessage) Bytes() []byte {
	bs := make([]byte, 0)
	buf := bytes.NewBuffer(bs)

	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(m.Length))
	buf.Write(lenBytes)

	if m.Type != MessageTypeKeepAlive {
		buf.Write([]byte{byte(m.Type)})
	}

	if len(m.Payload) > 0 {
		buf.Write(m.Payload)
	}

	return buf.Bytes()
}

/*
ReadMessage reads from a Reader, most likely net.Conn during real operation,
and decodes the next available message..
*/
func ReadMessage(r Reader) (m Message, err error) {
	buf := make([]byte, 4)
	log.Println("Waiting to read full")
	_, err = io.ReadFull(r, buf)
	if err != nil {
		log.Println("[ReadMessage] Error: ", err)
		return nil, err
	}
	mLen := int(binary.BigEndian.Uint32(buf))

	if mLen == 0 {
		return &BasicMessage{Length: 0, Type: MessageTypeKeepAlive}, nil
	}

	buf = make([]byte, mLen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		log.Println("[ReadMessage] Error: ", err)
		return nil, err
	}
	mType := MessageType(buf[0])

	if mLen > 1 {
		mPayload := buf[1:mLen]
		bm := BasicMessage{Length: mLen, Type: mType, Payload: mPayload}
		switch mType {
		case MessageTypeHave:
			pi := int(binary.BigEndian.Uint32(mPayload))
			m = &HaveMessage{BasicMessage: bm, PieceIndex: pi}
		case MessageTypeBitField:
			bf := BitFieldFromHexString(string(mPayload))
			m = &BitFieldMessage{BasicMessage: bm, BitField: bf}
		case MessageTypeRequest:
			pieceIndex := int(binary.BigEndian.Uint32(mPayload[0:4]))
			beginOffset := int(binary.BigEndian.Uint32(mPayload[4:8]))
			pieceLength := int(binary.BigEndian.Uint32(mPayload[8:12]))
			m = &RequestMessage{BasicMessage: bm, PieceIndex: pieceIndex, BeginOffset: beginOffset, PieceLength: pieceLength}
		case MessageTypePiece:
			pieceIndex := int(binary.BigEndian.Uint32(mPayload[0:4]))
			beginOffset := int(binary.BigEndian.Uint32(mPayload[4:8]))
			block := mPayload[8:]
			m = &PieceMessage{BasicMessage: bm, PieceIndex: pieceIndex, BeginOffset: beginOffset, Block: block}
		case MessageTypeCancel:
			pieceIndex := int(binary.BigEndian.Uint32(mPayload[0:4]))
			beginOffset := int(binary.BigEndian.Uint32(mPayload[4:8]))
			pieceLength := int(binary.BigEndian.Uint32(mPayload[8:12]))
			m = &CancelMessage{BasicMessage: bm, PieceIndex: pieceIndex, BeginOffset: beginOffset, PieceLength: pieceLength}
		case MessageTypePort:
			port := int(binary.BigEndian.Uint16(mPayload))
			m = &PortMessage{BasicMessage: bm, Port: port}
		default:
			m = &bm
		}
	} else {
		m = &BasicMessage{Length: mLen, Type: mType}
	}

	return
}
