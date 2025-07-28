package main

import (
	"dwarfserver/pkg/dwarf"
	"flag"
)

func main() {
	elfPath := flag.String("f", "", "elf path")
	flag.Parse()

	//dwarf.ShowMembers(*elfPath)
	dwarf.ShowStructTypedefs(*elfPath)

}
