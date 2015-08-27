package main

import (
	"fmt"
	"github.com/kr/pretty"
	"github.com/stratospark/torro/bencoding"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	println("TORRO!")

	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	} else {
		filename = "testfiles/TheInternetsOwnBoyTheStoryOfAaronSwartz_archive.torrent"

	}

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

	conv := func(val interface{}) string {
		b, _ := val.([]uint8)
		return string(b)
	}

	pretty.Println("Announce: ", conv(result["announce"]))
	pretty.Println("Announce-List", conv(result["annnounce-list"]))

	creationDate := int64(result["creation date"].(int))
	t := time.Unix(creationDate, 0)
	pretty.Println("Creation Date:", t.String())

	pretty.Println("Comment:", conv(result["comment"]))
	pretty.Println("Created by:", conv(result["created by"]))
	pretty.Println("Encoding:", conv(result["encoding"]))

	info := result["info"].(map[string]interface{})
	pretty.Println("Info Piece Length:", info["piece length"])
	pretty.Println("Info Private:", info["private"])

	pretty.Println("Info/Name:", conv(info["name"]))
	pretty.Println("Info/piece length:", info["piece length"])
	pretty.Println("Info/pieces:", len(conv(info["pieces"])))
	pretty.Println("Info/pieces/20:", len(conv(info["pieces"]))/20)
	//	pretty.Println("Info/pieces:", info["pieces"].(string)[:4])
	pretty.Println("Info/Length:", info["length"])
	pretty.Println("Info/md5sum:", info["md5sum"])
	//	pretty.Println("Info/files:", info["files"])

	if info["files"] != nil {
		files := info["files"].([]interface{})
		for i, val := range files {
			//		pretty.Println(i, ": ", val)
			pretty.Println("\n\n", i)
			file := val.(map[string]interface{})
			for key, val2 := range file {
				switch val2.(type) {
				case int:
					pretty.Println(key, ": ", val2)
				case []uint8:
					pretty.Println(key, ": ", conv(val2))
				case []interface{}:
					path := val2.([]interface{})
					pretty.Println("Paths:")
					for _, val3 := range path {
						pretty.Println("\t\t", conv(val3))
					}
				default:
					pretty.Println(key, ": ", val2)
				}
			}
		}
	}
}
