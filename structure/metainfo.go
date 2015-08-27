package structure

import (
	"github.com/stratospark/torro/bencoding"
	"io/ioutil"
	"time"
	"fmt"
)

type File struct {
	Length int
	MD5sum string
	Path   string
}

type InfoMode int

const (
	InfoModeSingle InfoMode = iota
	InfoModeMultiple
)

type Info struct {
	Mode        InfoMode
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
	if result["info"] != nil {
		addInfoFields(metainfo, result["info"].(map[string]interface{}))
	} else {
		panic("MISSING REQUIRED FIELD: info")
	}

	if result["announce"] != nil {
		b, _ := result["announce"].([]uint8)
		metainfo.Announce = string(b)
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
		b, _ := result["comment"].([]uint8)
		metainfo.Comment = string(b)
	}

	if result["created by"] != nil {
		b, _ := result["created by"].([]uint8)
		metainfo.CreatedBy = string(b)
	}

	if result["encoding"] != nil {
		b, _ := result["encoding"].([]uint8)
		metainfo.Encoding = string(b)
	}

	return metainfo
}

func addInfoFields(metainfo *Metainfo, infoMap map[string]interface{}) {
	info := &Info{}

	if infoMap["piece length"] != nil {
		info.PieceLength = infoMap["piece length"].(int)
	} else {
		panic("MISSING REQUIRED FIELD: piece length")
	}

	fmt.Println("ok")

	if infoMap["pieces"] != nil {
		b, _ := infoMap["pieces"].([]uint8)
		info.Pieces = string(b)
	} else {
		panic("MISSING REQUIRED FIELD: pieces")
	}

	if infoMap["private"] != nil {
		info.Private = infoMap["private"].(bool)
	}

	if infoMap["name"] != nil {
		b, _ := infoMap["name"].([]uint8)
		info.Name = string(b)
	} else {
		panic("MISSING REQUIRED FIELD: name")
	}

	// Check whether single or multiple file mode
	if infoMap["files"] != nil {
		info.Mode = InfoModeMultiple
	} else {
		info.Mode = InfoModeSingle

		if infoMap["length"] != nil {
			info.Length = infoMap["length"].(int)
		} else {
			panic("MISSING REQUIRED FIELD: length")
		}

		if infoMap["md5sum"] != nil {
			b, _ := infoMap["md5sum"].([]uint8)
			info.MD5Sum = string(b)
		}
	}

	metainfo.Info = *info
}
