package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/machinebox/sdk-go/textbox"
	"github.com/machinebox/twitterfeed"
)

func main() {
	var (
		addr = flag.String("addr", ":9000", "address")
	)

	consumerKey := os.Getenv("MOOD_COSUMER_KEY")
	consumerSecret := os.Getenv("MOOD_CONSUMER_SECRET")
	accessToken := os.Getenv("MOOD_ACCESS_TOKEN")
	accessSecret := os.Getenv("MOOD_ACCESS_SECRET")

	flag.Parse()
	tweetReader := twitterfeed.NewTweetReader(
		consumerKey,
		consumerSecret,
		accessToken,
		accessSecret)
	textbox := textbox.New("http://localhost:8080")
	fmt.Println(`mood by Machine Box - https://machinebox.io/
`)
	fmt.Println("Go to:", *addr+"...")
	srv := NewServer("./assets", tweetReader, textbox)
	if err := http.ListenAndServe(*addr, srv); err != nil {
		log.Fatalln(err)
	}
}
