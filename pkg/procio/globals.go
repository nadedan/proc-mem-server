package procio

import (
	"debug/elf"
	"fmt"
)

// variables is a map where the key is name and the value is the variable info
type variables map[string]variable

// variable represents a variable's name, address, and section.
type variable struct {
	Name          string
	Address       uint64
	SectionOffset uint64
	Size          uint64
	Section       string
	SectionAddr   uint64
}

// GlobalVariables extracts global variables from an ELF file.
func GlobalVariables(f *elf.File) (variables, error) {
	// Get the symbol table (try .symtab first, fall back to .dynsym).
	syms, err := f.Symbols()
	if err != nil {
		// If .symtab is not found, try .dynsym.
		syms, err = f.DynamicSymbols()
		if err != nil {
			return nil, fmt.Errorf("failed to read symbol table: %w", err)
		}
	}

	// Get section headers to map symbol section indices to section names.
	sections := f.Sections

	globalVars := variables{}

	// Iterate through symbols.
	for _, sym := range syms {
		// Filter for global variables:
		// - Type: STT_OBJECT (variables)
		if elf.ST_TYPE(sym.Info) != elf.STT_OBJECT {
			continue
		}
		// - Binding: STB_GLOBAL (global scope)
		if elf.ST_BIND(sym.Info) != elf.STB_GLOBAL {
			continue
		}
		if sym.Name[0:2] == "__" {
			continue
		}

		// Get the section index.
		sectionIdx := sym.Section
		sectionName := "unknown"

		// Map section index to section name.
		if sectionIdx < elf.SHN_LORESERVE && int(sectionIdx) < len(sections) {
			sectionName = sections[sectionIdx].Name
		}

		// Only include variables in .data or .bss sections.
		if sectionName == ".data" || sectionName == ".bss" {
			globalVars[sym.Name] = variable{
				Name:          sym.Name,
				Address:       sym.Value,
				Size:          sym.Size,
				Section:       sectionName,
				SectionAddr:   sections[sectionIdx].Addr,
				SectionOffset: sym.Value - sections[sectionIdx].Addr,
			}
		}
	}

	return globalVars, nil
}
