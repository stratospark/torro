package structure

import (
	"encoding/binary"
	"fmt"
	"github.com/stratospark/torro/bencoding"
	"net"
	"strconv"
	"strings"
)

type TrackerRequest struct {
	Metainfo   *Metainfo
	InfoHash   string
	PeerID     string
	Port       int
	Uploaded   int
	Downloaded int
	Compact    bool
	NoPeerID   bool
	Event      string
	IP         string
	NumWant    int
	Key        string
	TrackerID  string
}

func NewTrackerRequest(metainfo *Metainfo) *TrackerRequest {
	return &TrackerRequest{
		Metainfo: metainfo,
		InfoHash: metainfo.Info.Hash,
		PeerID:   "-TR2840-nj5ovtkoz2ed8", // TODO: generate unique PeerID
	}
}

func (request *TrackerRequest) Left() int {
	return request.Metainfo.Info.TotalBytes - request.Downloaded
}

func Btos(b bool) string {
	result := "0"
	if b {
		result = "1"
	}
	return result
}

func (request *TrackerRequest) GetURL() string {
	// TODO: handle AnnounceLists

	url := request.Metainfo.Announce
	url += "?info_hash=" + request.Metainfo.Info.Hash +
		"&peer_id=" + request.PeerID +
		"&port=" + strconv.Itoa(request.Port) +
		"&uploaded=" + strconv.Itoa(request.Uploaded) +
		"&downloaded=" + strconv.Itoa(request.Downloaded) +
		"&left=" + strconv.Itoa(request.Left()) +
		"&compact=" + Btos(request.Compact) +
		"&no_peer_id=" + Btos(request.NoPeerID) +
		"&event=" + request.Event

	return url
}

type Peer struct {
	IP   net.IP
	Port uint16
}

func (peer *Peer) String() string {
	return fmt.Sprintf(" (%s:%d) ", peer.IP, peer.Port)
}

type TrackerResponse struct {
	Complete    int
	Incomplete  int
	Downloaded  int
	Interval    int
	MinInterval int
	Peers       []Peer
}

func (tr *TrackerResponse) String() string {
	peerList := make([]string, 0)
	for _, peer := range tr.Peers {
		peerList = append(peerList, peer.String())
	}
	joinedPeers := strings.Join(peerList, ", ")

	return fmt.Sprintf("Response [Complete: %d, Incomplete %d, Downloaded: %d, Interval: %d, MinInterval: %d, Peers: %q]",
		tr.Complete, tr.Incomplete, tr.Downloaded, tr.Interval, tr.MinInterval, joinedPeers)
}

func NewTrackerResponse(responseStr string) *TrackerResponse {
	lex := bencoding.BeginLexing("response", responseStr, bencoding.LexBegin)
	tokens := bencoding.Collect(lex)
	parser := bencoding.Parse(tokens)
	o := parser.Output.(map[string]interface{})
	//	fmt.Println(o)

	// TODO: Handle required/optional fields
	response := &TrackerResponse{}
	addIntField(&response.Complete, o["complete"], true)
	addIntField(&response.Incomplete, o["incomplete"], true)
	addIntField(&response.Downloaded, o["downloaded"], false)
	addIntField(&response.Interval, o["interval"], true)
	addIntField(&response.MinInterval, o["min interval"], false)

	// TODO: Handle dictionary vs binary peer models
	peers := make([]Peer, 0)
	peerBytes := o["peers"].([]byte)
	for i := 0; i < len(peerBytes); i += 6 {
		ip := net.IPv4(peerBytes[i+3], peerBytes[i+2], peerBytes[i+1], peerBytes[i])

		port := binary.BigEndian.Uint16(peerBytes[i+4 : i+6])
		peers = append(peers, Peer{IP: ip, Port: port})
	}
	response.Peers = peers

	return response
}
