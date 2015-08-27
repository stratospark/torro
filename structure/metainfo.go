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

func addStringField(s *string, val interface{}, required bool) {
		if val != nil {
			b, _ := val.([]uint8)
			*s = string(b)
		} else {
			if required {
				panic("MISSING REQUIRED FIELD: announce")
			}
		}
}

func addIntField(s *int, val interface{}, required bool) {
		if val != nil {
			*s = val.(int)
		} else {
			if required {
				panic("MISSING REQUIRED FIELD: announce")
			}
		}
}

func addBoolField(s *bool, val interface{}, required bool) {
	if val != nil {
		*s = val.(bool)
	} else {
		if required {
			panic("MISSING REQUIRED FIELD: announce")
		}
	}
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

	addStringField(&metainfo.Announce, result["announce"], true)

	// Optional fields
	if result["announce-list"] != nil {
		// TODO
	}

	if result["creation date"] != nil {
		creationDate := int64(result["creation date"].(int))
		t := time.Unix(creationDate, 0)
		metainfo.CreationDate = t
	}

	addStringField(&metainfo.Comment, result["comment"], false)
	addStringField(&metainfo.CreatedBy, result["created by"], false)
	addStringField(&metainfo.Encoding, result["encoding"], false)

	return metainfo
}

func addInfoFields(metainfo *Metainfo, infoMap map[string]interface{}) {
	info := &Info{}

	addIntField(&info.PieceLength, infoMap["piece length"], true)

	// TODO: may need to keep this as byte array?
	addStringField(&info.Pieces, infoMap["pieces"], true)
	addBoolField(&info.Private, infoMap["private"], false)
	addStringField(&info.Name, infoMap["name"], true)

	// Check whether single or multiple file mode
	if infoMap["files"] != nil {
		info.Mode = InfoModeMultiple
		files := infoMap["files"].(map[string]interface{})
		for _, file := range files {
			fmt.Println(file)
		}
	} else {
		info.Mode = InfoModeSingle
		addIntField(&info.Length, infoMap["length"], true)
		addStringField(&info.MD5Sum, infoMap["md5sum"], false)
	}

	metainfo.Info = *info
}
