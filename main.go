package main

import (
	"github.com/stratospark/torro/bencoding"
	"io/ioutil"
	"github.com/kr/pretty"
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

	output := bencoding.Parse(tokens)

//	reuslt := output.Output.(map[string]interface{})

	pretty.Print(output.Output)
//	tokenCounts := make(map[bencoding.Token]int)
//	for _, t := range tokens {
//		tokenCounts[t]++
//		fmt.Println(t)
//	}
//	fmt.Println(tokenCounts)
}
