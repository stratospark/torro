package main

import (
	"flag"
	"fmt"
	"github.com/NebulousLabs/go-upnp"
	"github.com/kr/pretty"
	"github.com/stratospark/torro/bencoding"
	"github.com/stratospark/torro/client"
	"github.com/stratospark/torro/structure"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	// Set up logging to file and stdout
	f, err := os.OpenFile("torro.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Error opening file: %v", err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	println("TORRO!\n\n\n")

	d, _ := upnp.Discover()
	ip, _ := d.ExternalIP()
	_ = d.Forward(55555, "torro")
	defer d.Clear(55555)
	log.Printf("Discovered: %q\n", d)
	log.Printf("External IP: %q\n", ip)
	log.Printf("Location: %q\n", d.Location())

	// Read command line flags and arguments
	pPrint := flag.String("print", "metainfo", "either tokens, parsed, or metainfo")
	flag.Parse()

	var filename string
	fmt.Println(flag.Args())
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	} else {
		filename = "testfiles/TheInternetsOwnBoyTheStoryOfAaronSwartz_archive.torrent"
	}

	// Read actual .torrent file
	fmt.Println("Parsing: ", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	// Lex and Parse .torrent file
	torrentStr := string(data)
	lex := bencoding.BeginLexing(".torrent", torrentStr, bencoding.LexBegin)
	tokens := bencoding.Collect(lex)

	output := bencoding.Parse(tokens)
	result := output.Output.(map[string]interface{})

	// Read .torrent metainfo and make request to the announce URL
	metainfo := structure.NewMetainfo(filename)

	c := client.NewTrackerClient()
	req := structure.NewTrackerRequest(metainfo)
	req.PeerID = "-qB3230-u~QGMmUs~yXH"
	req.Port = 55555
	req.Compact = true
	req.NoPeerID = true
	res, err := c.MakeAnnounceRequest(req, client.TrackerRequestStarted)
	if err != nil {
		fmt.Println(res.FailureReason)
		panic(err.Error())
	}
	fmt.Println(res)

	log.Println("StartListening")
	port := 55555
	s := client.NewBTService(port)
	s.StartListening()

	switch *pPrint {
	case "tokens":
		PrintTokens(&tokens)
	case "parsed":
		PrintParsedStructure(result)
	case "metainfo":
		PrintMetainfo(metainfo)
	default:
		PrintMetainfo(metainfo)
	}

	time.Sleep(time.Second * 60)

}

func PrintTokens(tokens *[]bencoding.Token) {
	fmt.Println("Printing Tokens:\n\n")
	for _, token := range *tokens {
		fmt.Println(token)
	}
}

func PrintParsedStructure(result map[string]interface{}) {

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

	//	if info["files"] != nil {
	//		files := info["files"].([]interface{})
	//		for i, val := range files {
	//			//		pretty.Println(i, ": ", val)
	//			pretty.Println("\n\n", i)
	//			file := val.(map[string]interface{})
	//			for key, val2 := range file {
	//				switch val2.(type) {
	//				case int:
	//					pretty.Println(key, ": ", val2)
	//				case []uint8:
	//					pretty.Println(key, ": ", conv(val2))
	//				case []interface{}:
	//					path := val2.([]interface{})
	//					pretty.Println("Paths:")
	//					for _, val3 := range path {
	//						pretty.Println("\t\t", conv(val3))
	//					}
	//				default:
	//					pretty.Println(key, ": ", val2)
	//				}
	//			}
	//		}
	//	}
}

func PrintMetainfo(metainfo *structure.Metainfo) {
	fmt.Println(metainfo.Announce)
}
