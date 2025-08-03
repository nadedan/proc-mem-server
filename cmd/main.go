package main

import (
	"dwarfserver/pkg/procio"
	"flag"
)

func main() {
	elfPath := flag.String("f", "", "elf path")
	flag.Parse()

	procio.ShowGlobals(*elfPath)
	//dwarf.ShowStructTypedefs(*elfPath)

}
