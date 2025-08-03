package main

import (
	"flag"
	"fmt"

	"dwarfserver/pkg/procio"
)

func main() {
	binName := flag.String("b", "", "binName")
	flag.Parse()

	procMmap, err := procio.NewMmap(*binName)
	if err != nil {
		panic(err)
	}

	outerInt := procMmap.Int32(0x48)
	fmt.Printf("outerInt: %d\n", outerInt)

}
