package dwarf

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
)

type StructField struct {
	Name   string
	Offset uint64
	Size   uint64
	Type   dwarf.Type
}

func ShowStructTypedefs(fileName string) {

	file, err := elf.Open(fileName)
	if err != nil {
		panic(err)
	}

	dwarfData, err := file.DWARF()
	if err != nil {
		panic(err)
	}

	tds := structTypedefs(dwarfData)
	for name, td := range tds {
		fmt.Printf("typedef: %s\n", name)
		structType := td.Type.(*dwarf.StructType)
		for _, field := range structType.Field {
			if field.Type == nil {
				continue
			}
			fmt.Printf("  field: %s\n", field.Name)
			fmt.Printf("    type  : %s\n", field.Type)
			fmt.Printf("    offset: %d\n", field.ByteOffset)
			fmt.Printf("    size  : %d\n", field.Type.Size())
		}
	}

}

type StructDefs map[string]*dwarf.TypedefType

func structTypedefs(dwarfData *dwarf.Data) StructDefs {
	m := make(map[string]*dwarf.TypedefType)

	reader := dwarfData.Reader()
	for {
		entry, err := reader.Next()
		if entry == nil || err != nil {
			break
		}

		name, _ := entry.Val(dwarf.AttrName).(string)
		if name == "" {
			continue
		}

		if entry.Tag != dwarf.TagTypedef {
			continue
		}

		typEntry, err := dwarfData.Type(entry.Offset)
		if err != nil {
			continue
		}

		typedefType, ok := typEntry.(*dwarf.TypedefType)
		if !ok {
			continue
		}
		_, ok = typedefType.Type.(*dwarf.StructType)
		if !ok {
			continue
		}

		m[typedefType.Name] = typedefType
	}

	return m
}

func ShowGlobals(fileName string) {

	file, err := elf.Open(fileName)
	if err != nil {
		panic(err)
	}

	dwarfData, err := file.DWARF()
	if err != nil {
		panic(err)
	}

	structs := structTypedefs(dwarfData)

	reader := dwarfData.Reader()
	results := []StructField{}

	for {
		entry, err := reader.Next()
		if entry == nil || err != nil {
			break
		}

		name, _ := entry.Val(dwarf.AttrName).(string)
		if name == "" {
			continue
		}

		if entry.Tag != dwarf.TagVariable {
			continue
		}

		typeOffset, ok := entry.Val(dwarf.AttrType).(dwarf.Offset)
		if !ok {
			continue
		}

		typEntry, err := dwarfData.Type(typeOffset)
		if err != nil {
			fmt.Printf("could not get typEntry\n")
			continue
		}

		structType, ok := typEntry.(*dwarf.StructType)
		if !ok {
			typedefType, ok := typEntry.(*dwarf.TypedefType)
			if !ok {
				continue
			}
			structType = typedefType.Type.(*dwarf.StructType)
		}

		members := collectStructMembers(structType, name, 0, structs)
		results = append(results, members...)
	}

	for _, info := range results {
		fmt.Printf("Member: %s, Type: %s, Offset: %d, Size: %d\n", info.Name, info.Type, info.Offset, info.Size)
	}
}

func collectStructMembers(structType *dwarf.StructType, path string, offset uint64, structs StructDefs) []StructField {
	members := []StructField{}

	for _, field := range structType.Field {
		if field.Type == nil {
			continue
		}
		member := StructField{
			Name:   path + "." + field.Name,
			Offset: offset + uint64(field.ByteOffset),
			Size:   uint64(field.Type.Size()),
			Type:   field.Type,
		}
		_, isSubStruct := structs[field.Type.Common().Name]
		if !isSubStruct {
			members = append(members, member)
			continue
		}
		members = append(members,
			collectStructMembers(
				field.Type.(*dwarf.TypedefType).Type.(*dwarf.StructType),
				path+"."+field.Name,
				offset+uint64(field.ByteOffset),
				structs,
			)...,
		)
	}
	return members
}
