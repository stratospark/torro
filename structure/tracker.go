package structure

import (
	"strconv"
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
