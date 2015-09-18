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

	// Get the rest of the handshake message
	buf = make([]byte, pstrLen+48)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		// Fewer bytes than expected?
		log.Printf("[ReadHandshake] More ReadFull Error: %s, %x, %d", err, buf, len(buf))
		return nil, err
	}

	name := string(buf[0:pstrLen])
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

func (h *Handshake) GetType() MessageType {
	return MessageTypeHandshake
}

type MessageType byte

const (
	MessageTypeHandshake     MessageType = 254
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
	GetType() MessageType
}

type BasicMessage struct {
	Length  int
	Type    MessageType
	Payload []byte
}

func (bm *BasicMessage) GetType() MessageType {
	return bm.Type
}

type KeepAliveMessage struct {
	BasicMessage
}

func NewKeepAliveMessage() *KeepAliveMessage {
	msg := &KeepAliveMessage{BasicMessage: BasicMessage{Type: MessageTypeKeepAlive, Length: 0}}
	return msg
}

type ChokeMessage struct {
	BasicMessage
}

func NewChokeMessage() *ChokeMessage {
	msg := &ChokeMessage{BasicMessage: BasicMessage{Type: MessageTypeChoke, Length: 1}}
	return msg
}

type UnchokeMessage struct {
	BasicMessage
}

func NewUnchokeMessage() *UnchokeMessage {
	msg := &UnchokeMessage{BasicMessage: BasicMessage{Type: MessageTypeUnchoke, Length: 1}}
	return msg
}

type InterestedMessage struct {
	BasicMessage
}

func NewInterestedMessage() *InterestedMessage {
	msg := &InterestedMessage{BasicMessage: BasicMessage{Type: MessageTypeInterested, Length: 1}}
	return msg
}

type NotInterestedMessage struct {
	BasicMessage
}

func NewNotInterestedMessage() *NotInterestedMessage {
	msg := &NotInterestedMessage{BasicMessage: BasicMessage{Type: MessageTypeNotInterested, Length: 1}}
	return msg
}

type HaveMessage struct {
	BasicMessage
	PieceIndex int
}

func NewHaveMessage(pieceIndex int) *HaveMessage {
	msg := &HaveMessage{BasicMessage: BasicMessage{Type: MessageTypeHave, Length: 5}, PieceIndex: pieceIndex}
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(pieceIndex))
	msg.Payload = bs
	return msg
}

type BitFieldMessage struct {
	BasicMessage
	BitField *BitField
}

func NewBitFieldMessage(bf *BitField) *BitFieldMessage {
	msg := &BitFieldMessage{BasicMessage: BasicMessage{Type: MessageTypeBitField, Length: 5}, BitField: bf}
	msg.Payload = bf.Bytes()
	return msg
}

type RequestMessage struct {
	BasicMessage
	PieceIndex  int
	BeginOffset int
	PieceLength int
}

func NewRequestMessage(pieceIndex, beginOffset, pieceLength int) *RequestMessage {
	msg := &RequestMessage{BasicMessage: BasicMessage{Type: MessageTypeRequest, Length: 13}, PieceIndex: pieceIndex, BeginOffset: beginOffset, PieceLength: pieceLength}
	buf := bytes.NewBuffer(make([]byte, 0))
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(pieceIndex))
	buf.Write(bs)
	binary.BigEndian.PutUint32(bs, uint32(beginOffset))
	buf.Write(bs)
	binary.BigEndian.PutUint32(bs, uint32(pieceLength))
	buf.Write(bs)
	msg.Payload = buf.Bytes()
	return msg
}

type PieceMessage struct {
	BasicMessage
	PieceIndex  int
	BeginOffset int
	Block       []byte
}

func NewPieceMessage(pieceIndex int, beginOffset int, block []byte) *PieceMessage {
	msg := &PieceMessage{BasicMessage: BasicMessage{Type: MessageTypePiece, Length: 13}, PieceIndex: pieceIndex, BeginOffset: beginOffset, Block: block}
	buf := bytes.NewBuffer(make([]byte, 0))
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(pieceIndex))
	buf.Write(bs)
	binary.BigEndian.PutUint32(bs, uint32(beginOffset))
	buf.Write(bs)
	buf.Write(block)
	msg.Payload = buf.Bytes()
	return msg
}

type CancelMessage struct {
	BasicMessage
	PieceIndex  int
	BeginOffset int
	PieceLength int
}

func NewCancelMessage(pieceIndex, beginOffset, pieceLength int) *CancelMessage {
	msg := &CancelMessage{BasicMessage: BasicMessage{Type: MessageTypeCancel, Length: 13}, PieceIndex: pieceIndex, BeginOffset: beginOffset, PieceLength: pieceLength}
	buf := bytes.NewBuffer(make([]byte, 0))
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(pieceIndex))
	buf.Write(bs)
	binary.BigEndian.PutUint32(bs, uint32(beginOffset))
	buf.Write(bs)
	binary.BigEndian.PutUint32(bs, uint32(pieceLength))
	buf.Write(bs)
	msg.Payload = buf.Bytes()
	return msg
}

type PortMessage struct {
	BasicMessage
	Port int
}

func NewPortMessage(port int) *PortMessage {
	msg := &PortMessage{BasicMessage: BasicMessage{Type: MessageTypePort, Length: 3}, Port: port}
	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, uint16(port))
	msg.Payload = bs
	return msg
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
		return &KeepAliveMessage{BasicMessage: BasicMessage{Length: 0, Type: MessageTypeKeepAlive}}, nil
	}

	buf = make([]byte, mLen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		log.Println("[ReadMessage] Error: ", err)
		return nil, err
	}
	mType := MessageType(buf[0])

	if mLen >= 1 {
		mPayload := buf[1:mLen]
		bm := BasicMessage{Length: mLen, Type: mType, Payload: mPayload}
		switch mType {
		case MessageTypeChoke:
			bm.Payload = nil
			m = &ChokeMessage{BasicMessage: bm}
		case MessageTypeUnchoke:
			bm.Payload = nil
			m = &UnchokeMessage{BasicMessage: bm}
		case MessageTypeInterested:
			bm.Payload = nil
			m = &InterestedMessage{BasicMessage: bm}
		case MessageTypeNotInterested:
			bm.Payload = nil
			m = &NotInterestedMessage{BasicMessage: bm}
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
		}
	} else {
		m = &BasicMessage{Length: mLen, Type: mType}
	}

	return
}
