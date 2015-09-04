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

type Reader interface {
	Read(p []byte) (n int, err error)
}

func NewHandshake(r Reader) (h *Handshake, err error) {
	buf := make([]byte, 1)
	log.Println("Waiting to readfull")
	_, err = io.ReadFull(r, buf)
	if err != nil {
		log.Println("[HandleConnection] Error: ", err)
		return nil, err
	}
	pstrLen := int(buf[0])

	// Get the rest of the handshake message
	buf = make([]byte, pstrLen+48)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		// Fewer bytes than expected?
		log.Println("[HandleConnection] Error: ", err)
		return nil, err
	}

	name := string(buf[0:pstrLen])
	if name != "BitTorrent protocol" {
		log.Println("[HandleConnection] Not BitTorrent protocol handshake")
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

	return h, nil
}

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

func (m BasicMessage) Bytes() []byte {
	bs := make([]byte, 0)
	buf := bytes.NewBuffer(bs)

	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(m.Length))
	buf.Write(lenBytes)

	buf.Write([]byte{byte(m.Type)})
	return buf.Bytes()
}

func (m HaveMessage) Bytes() []byte {
	return []byte("\x99")
}

func (m BitFieldMessage) Bytes() []byte {
	return []byte("\x99")
}

func ReadMessage(r Reader) (m Message, err error) {
	buf := make([]byte, 4)
	log.Println("Waiting to read full")
	_, err = io.ReadFull(r, buf)
	if err != nil {
		log.Println("[HandleConnection] Error: ", err)
		return nil, err
	}
	mLen := int(binary.BigEndian.Uint32(buf))

	if mLen == 0 {
		return &BasicMessage{Length: 0, Type: MessageTypeKeepAlive}, nil
	}

	buf = make([]byte, mLen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		log.Println("[HandleConnection] Error: ", err)
		return nil, err
	}
	mType := MessageType(buf[0])

	if mLen > 1 {
		mPayload := buf[1:mLen]
		switch mType {
		case MessageTypeHave:
			pi := int(binary.BigEndian.Uint32(mPayload))
			m = &HaveMessage{BasicMessage: BasicMessage{Length: mLen, Type: mType, Payload: mPayload}, PieceIndex: pi}
		case MessageTypeBitField:
			bf := BitFieldFromHexString(string(mPayload))
			m = &BitFieldMessage{BasicMessage: BasicMessage{Length: mLen, Type: mType, Payload: mPayload}, BitField: bf}
		default:
			m = &BasicMessage{Length: mLen, Type: mType, Payload: mPayload}
		}
	} else {
		m = &BasicMessage{Length: mLen, Type: mType}
	}

	return
}
