package main
import (
	"io/ioutil"
	"github.com/stratospark/torro/bencoding"
	"fmt"
)

func main() {
	println("TORRO!")

	data, err := ioutil.ReadFile("ubuntu.torrent")
	if err != nil {
		panic(err)
	}

	torrentStr := string(data)
	lex := bencoding.BeginLexing(".torrent", torrentStr, bencoding.LexBegin)
	tokens := bencoding.Collect(lex)
	for _, t := range tokens {
		fmt.Println(t)
	}
}

