package main

import (
	"fmt"
	"github.com/kr/pretty"
	"github.com/stratospark/torro/bencoding"
	"io/ioutil"
	"time"
)

func main() {
	println("TORRO!")

	//	filename := "testfiles/ubuntu.torrent"
	filename := "testfiles/TheInternetsOwnBoyTheStoryOfAaronSwartz_archive.torrent"
	fmt.Println("Parsing: ", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	torrentStr := string(data)
	lex := bencoding.BeginLexing(".torrent", torrentStr, bencoding.LexBegin)
	tokens := bencoding.Collect(lex)

	output := bencoding.Parse(tokens)
	result := output.Output.(map[string]interface{})

	pretty.Println("Announce: ", result["announce"])
	pretty.Println("Announce-List", result["annnounce-list"])

	creationDate := int64(result["creation date"].(int))
	t := time.Unix(creationDate, 0)
	pretty.Println("Creation Date:", t.String())

	pretty.Println("Comment:", result["comment"])
	pretty.Println("Created by:", result["created by"])
	pretty.Println("Encoding:", result["encoding"])

	info := result["info"].(map[string]interface{})
	pretty.Println("Info Piece Length:", info["piece length"])
	pretty.Println("Info Private:", info["private"])

	pretty.Println("Info/Name:", info["name"])
	pretty.Println("Info/piece length:", info["piece length"])
	pretty.Println("Info/pieces:", len(info["pieces"].(string)))
	pretty.Println("Info/pieces/20:", len(info["pieces"].(string))/20)
	//	pretty.Println("Info/pieces:", info["pieces"].(string)[:4])
	pretty.Println("Info/Length:", info["length"])
	pretty.Println("Info/md5sum:", info["md5sum"])
	pretty.Println("Info/files:", info["files"])
}
