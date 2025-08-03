package main

import (
	"debug/elf"
	"flag"
	"fmt"

	"dwarfserver/pkg/procio"
)

func main() {
	elfPath := flag.String("f", "", "elf path")
	flag.Parse()

	file, err := elf.Open(*elfPath)
	if err != nil {
		panic(err)
	}

	gvs, err := procio.GlobalVariables(file)
	if err != nil {
		panic(err)
	}

	reqSections := procio.RequiredSections(gvs)
	for name, section := range reqSections {
		fmt.Printf("Name: %s, Addr: 0x%0X, Size: %d\n", name, section.Addr(), section.Size())
	}
}
