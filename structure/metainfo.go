package structure

import (
	"github.com/stratospark/torro/bencoding"
	"io/ioutil"
	"time"
)

type File struct {
	Length int
	MD5sum string
	Path   string
}

type Info struct {
	PieceLength int
	Pieces      string
	Private     bool
	Name        string
	Length      int
	MD5Sum      string
	Files       []File
}

type Metainfo struct {
	Info         Info
	Announce     string
	AnnounceList [][]string
	CreationDate time.Time
	Comment      string
	CreatedBy    string
	Encoding     string
}

func NewMetainfo(filename string) *Metainfo {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	torrentStr := string(data)
	lex := bencoding.BeginLexing(".torrent", torrentStr, bencoding.LexBegin)
	tokens := bencoding.Collect(lex)

	output := bencoding.Parse(tokens)
	result := output.Output.(map[string]interface{})

	metainfo := &Metainfo{}

	// Required fields
	if result["announce"] != nil {
		metainfo.Announce = result["announce"].(string)
	} else {
		panic("MISSING REQUIRED FIELD: announce")
	}

	// Optional fields
	if result["announce-list"] != nil {
		// TODO
	}

	if result["creation date"] != nil {
		creationDate := int64(result["creation date"].(int))
		t := time.Unix(creationDate, 0)
		metainfo.CreationDate = t
	}

	if result["comment"] != nil {
		metainfo.Comment = result["comment"].(string)
	}

	if result["created by"] != nil {
		metainfo.CreatedBy = result["created by"].(string)
	}

	if result["encoding"] != nil {
		metainfo.Encoding = result["encoding"].(string)
	}

	return metainfo
}
