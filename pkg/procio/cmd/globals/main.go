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
	for _, gv := range gvs {
		fmt.Printf("Name: %s, Addr: 0x%X, Size: %d, SectionAddr: 0x%X\n", gv.Name, gv.Address, gv.Size, gv.SectionAddr)
	}

}
