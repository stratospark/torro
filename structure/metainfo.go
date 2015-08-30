package structure

import (
	"crypto/sha1"
	"errors"
	"fmt"
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
	addIntField("length", &file.Length, rawFile["length"], true)
	addStringField("md5sum", &file.MD5sum, rawFile["md5sum"], false)
	addStringField("md5", &file.MD5sum, rawFile["md5"], false)

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
	TotalBytes  int
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

func addStringField(name string, s *string, val interface{}, required bool) error {
	if val != nil {
		b, _ := val.([]uint8)
		*s = string(b)
	} else {
		if required {
			return errors.New(fmt.Sprint("Missing Required Field: ", name))
		}
	}
	return nil
}

func addIntField(name string, s *int, val interface{}, required bool) error {
	if val != nil {
		*s = val.(int)
	} else {
		if required {
			return errors.New(fmt.Sprint("Missing Required Field: ", name))
		}
	}
	return nil
}

func addBoolField(name string, s *bool, val interface{}, required bool) error {
	if val != nil {
		*s = val.(bool)
	} else {
		if required {
			return errors.New(fmt.Sprint("Missing Required Field: ", name))
		}
	}
	return nil
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

	addStringField("announce", &metainfo.Announce, result["announce"], true)

	// Optional fields
	if result["announce-list"] != nil {
		// TODO
	}

	if result["creation date"] != nil {
		creationDate := int64(result["creation date"].(int))
		t := time.Unix(creationDate, 0)
		metainfo.CreationDate = t
	}

	addStringField("comment", &metainfo.Comment, result["comment"], false)
	addStringField("created by", &metainfo.CreatedBy, result["created by"], false)
	addStringField("encoding", &metainfo.Encoding, result["encoding"], false)

	return metainfo
}

func addInfoFields(metainfo *Metainfo, infoMap map[string]interface{}) {
	info := &Info{}

	addIntField("piece length", &info.PieceLength, infoMap["piece length"], true)

	// TODO: may need to keep this as byte array?
	addStringField("pieces", &info.Pieces, infoMap["pieces"], true)
	addBoolField("private", &info.Private, infoMap["private"], false)
	addStringField("name", &info.Name, infoMap["name"], true)

	totalBytes := 0

	// Check whether single or multiple file mode
	if infoMap["files"] != nil {
		info.Mode = InfoModeMultiple
		rawFiles := infoMap["files"].([]interface{})
		files := make([]File, 0)
		for _, rawFile := range rawFiles {
			newFile := *NewFile(rawFile)
			files = append(files, newFile)
			totalBytes += newFile.Length
		}
		info.Files = files
	} else {
		info.Mode = InfoModeSingle
		addIntField("length", &info.Length, infoMap["length"], true)
		addStringField("md5sum", &info.MD5Sum, infoMap["md5sum"], false)
		totalBytes = info.Length
	}

	info.TotalBytes = totalBytes

	metainfo.Info = *info
}
