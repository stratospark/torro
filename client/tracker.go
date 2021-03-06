package client

import (
	"github.com/stratospark/torro/structure"
	"io/ioutil"
	"log"
	"net/http"
)

type TrackerClient struct {
	HTTP *http.Client
}

type TrackerRequestEvent string

const (
	TrackerRequestStarted   TrackerRequestEvent = "started"
	TrackerRequestStopped   TrackerRequestEvent = "stopped"
	TrackerRequestCompleted TrackerRequestEvent = "completed"
)

func NewTrackerClient() *TrackerClient {
	tc := &TrackerClient{
		HTTP: http.DefaultClient,
	}

	return tc
}

func (tc *TrackerClient) MakeAnnounceRequest(req *structure.TrackerRequest, event TrackerRequestEvent) (tr *structure.TrackerResponse, err error) {
	// TODO: Validate that tracker request has valid fields, e.g. peer_id = 20 bytes
	req.Event = string(event)
	url := req.GetURL()
	resp, err := tc.HTTP.Get(url)
	log.Print("MakeAnounceRequest, URL: ", url)
	if err != nil {
		return tr, err
	}

	contents, err := ioutil.ReadAll(resp.Body)
	log.Println("MakeAnnounceRequest, Contents: ", string(contents))
	if err != nil {
		return tr, err
	}

	tr, err = structure.NewTrackerResponse(string(contents))
	if err != nil {
		return tr, err
	}
	log.Println("MakeAnnounceRequest, TrackerResponse: ", tr)

	return tr, nil
}
