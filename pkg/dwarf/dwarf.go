package dwarf

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
)

type Member struct {
	Path   string
	Offset uint64
	Size   uint64
}

func ShowMembers(fileName string) {
	file, err := elf.Open(fileName)
	if err != nil {
		panic(err)
	}

	dwarfData, err := file.DWARF()
	if err != nil {
		panic(err)
	}

	reader := dwarfData.Reader()
	results := []Member{}

	for {
		entry, err := reader.Next()
		if entry == nil || err != nil {
			break
		}

		if entry.Tag != dwarf.TagVariable {
			continue
		}

		name, _ := entry.Val(dwarf.AttrName).(string)
		if name == "" {
			continue
		}

		typOff, ok := entry.Val(dwarf.AttrType).(dwarf.Offset)
		if !ok {
			continue
		}

		typEntry, err := dwarfData.Type(typOff)
		if err != nil {
			continue
		}

		structType, ok := typEntry.(*dwarf.StructType)
		if !ok {
			continue
		}

		pathPrefix := name
		members := collectStructMembers(dwarfData, structType, pathPrefix)
		results = append(results, members...)
	}

	for _, info := range results {
		fmt.Printf("Member: %s, Offset: %d, Size: %d\n", info.Path, info.Offset, info.Size)
	}
}

func collectStructMembers(dwarfData *dwarf.Data, structType *dwarf.StructType, pathPrefix string) []Member {
	members := []Member{}

	for _, field := range structType.Field {
		if field.Type == nil {
			continue
		}
		fieldPath := pathPrefix + "." + field.Name
		member := Member{
			Path:   fieldPath,
			Offset: uint64(field.ByteOffset),
			Size:   uint64(field.Type.Size()),
		}
		members = append(members, member)

		if subStruct, ok := field.Type.(*dwarf.StructType); ok {
			subMembers := collectStructMembers(dwarfData, subStruct, fieldPath)
			members = append(members, subMembers...)
		}
	}
	return members
}
