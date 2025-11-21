package main

import (
	"encoding/binary"
	"flag"
	"fmt"

	"dwarfserver/pkg/procio"
)

func main() {
	binName := flag.String("b", "", "binName")
	flag.Parse()

	fmt.Printf("got bin name %s\n", *binName)
	pvm, err := procio.NewProcVm(*binName)
	if err != nil {
		panic(err)
	}

	b := make([]byte, 4)
	n, err := pvm.Read(0x20048, b)
	if err != nil {
		panic(err)
	}
	fmt.Printf("read %d bytes\n", n)

	outerInt := int32(binary.LittleEndian.Uint32(b))
	fmt.Printf("outerInt: %d\n", outerInt)

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, 13)
	n, err = pvm.Write(0x20048, b)
	if err != nil {
		panic(err)
	}
	fmt.Printf("wrote %d bytes\n", n)

	b = make([]byte, 4)
	n, err = pvm.Read(0x20048, b)
	if err != nil {
		panic(err)
	}
	fmt.Printf("read %d bytes\n", n)

	outerInt = int32(binary.LittleEndian.Uint32(b))
	fmt.Printf("outerInt: %d\n", outerInt)
}
