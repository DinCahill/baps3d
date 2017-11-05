// Package console is a simple console for inputting Bifrost commands.
package console

import (
	"fmt"
	"os"

	"github.com/UniversityRadioYork/baps3d/bifrost"
)

type Console struct {
	msgs chan bifrost.Message
	in   *bifrost.Tokeniser
}

func New(msgs chan bifrost.Message) *Console {
	return &Console{msgs: msgs, in: bifrost.NewTokeniser(os.Stdin)}
}

func (c *Console) RunRx() {
	for m := range c.msgs {
		mbytes, err := m.Pack()
		if err != nil {
			fmt.Println("-> rx error:", err)
			continue
		}
		os.Stdout.Write(mbytes)
	}
}

func (c *Console) RunTx() {
	for {
		line, terr := c.in.Tokenise()
		if terr != nil {
			fmt.Println("-> got error:", terr)
			return
		}

		msg, merr := bifrost.LineToMessage(line)
		if merr != nil {
			fmt.Println("-> invalid message:", merr)
			continue
		}

		c.msgs <- *msg
	}
}