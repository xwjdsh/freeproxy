package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
)

func main() {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tele.OnDocument, func(c tele.Context) error {
		fileName := c.Message().Document.FileName
		if err := c.Bot().Download(&c.Message().Document.File, fileName); err != nil {
			return err
		}
		return c.Send(fmt.Sprintf("[%s] downloaded", fileName))
	})

	b.Start()
}
