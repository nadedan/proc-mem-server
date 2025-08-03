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

	fields, err := procio.StructFields(file)
	if err != nil {
		panic(err)
	}
	for _, field := range fields {
		fmt.Printf("Field: %s %s, Type: %s, Offset: %d, Size: %d\n", field.BaseName, field.Path, field.Type, field.Offset, field.Size)
	}
}
