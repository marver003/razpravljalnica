package client

import (
	"flag"
	"fmt"
)

// TODO: CLI client, se poveže s serverjem, demonstracija programa se nahaja tuki

func main() {
	urlPtr := flag.String("u", "localhost", "server url")
	portPtr := flag.Int("p", 12345, "server port number")

	fmt.Printf("%s:%d", *urlPtr, *portPtr)
}
