package structure

type TrackerRequest struct {
	InfoHash   string
	PeerID     string
	Port       int
	Uploaded   int
	Downloaded int
	Left       int
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
		InfoHash: metainfo.Info.Hash,
		PeerID:   "-TR2840-nj5ovtkoz2ed8", // TODO: generate unique PeerID
	}
}
