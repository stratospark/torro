package structure

import (
	"errors"
	"fmt"
	"io"
	"log"
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
		return nil, errors.New("Not BitTorrent protocol handshake")
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
