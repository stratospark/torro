package structure

import (
	"crypto/sha1"
	"github.com/stratospark/torro/bencoding"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)

type File struct {
	Length int
	MD5sum string
	Path   string
}

func NewFile(f interface{}) *File {
	rawFile := f.(map[string]interface{})
	file := &File{}
	addIntField(&file.Length, rawFile["length"], true)
	addStringField(&file.MD5sum, rawFile["md5sum"], false)
	addStringField(&file.MD5sum, rawFile["md5"], false)

	paths := rawFile["path"].([]interface{})

	pathStrings := make([]string, 0)
	for _, path := range paths {
		b, _ := path.([]uint8)
		pathStrings = append(pathStrings, string(b))
	}
	fullPath := strings.Join(pathStrings, "/")
	file.Path = fullPath

	return file
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
	Hash        string
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

func getRightEncodedSHA1(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return strings.ToLower(url.QueryEscape(string(h.Sum(nil))))
}

func NewMetainfo(filename string) *Metainfo {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	torrentStr := string(data)
	lex := bencoding.BeginLexing(".torrent", torrentStr, bencoding.LexBegin)
	tokens := bencoding.Collect(lex)

	rawInfoVal := bencoding.GetBencodedInfo(tokens)

	output := bencoding.Parse(tokens)
	result := output.Output.(map[string]interface{})

	metainfo := &Metainfo{}

	// Required fields
	if result["info"] != nil {
		addInfoFields(metainfo, result["info"].(map[string]interface{}))
		metainfo.Info.Hash = getRightEncodedSHA1(rawInfoVal)
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
		rawFiles := infoMap["files"].([]interface{})
		files := make([]File, 0)
		for _, rawFile := range rawFiles {
			files = append(files, *NewFile(rawFile))
		}
		info.Files = files
	} else {
		info.Mode = InfoModeSingle
		addIntField(&info.Length, infoMap["length"], true)
		addStringField(&info.MD5Sum, infoMap["md5sum"], false)
	}

	metainfo.Info = *info
}
