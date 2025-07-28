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

func structTypedefs(dwarfData *dwarf.Data) map[string]*dwarf.TypedefType {
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

		name, _ := entry.Val(dwarf.AttrName).(string)
		if name == "" {
			continue
		}

		fmt.Printf("name: %s\n", name)
		fmt.Printf("entry.Tag: %T %+v\n", entry.Tag, entry.Tag)
		if entry.Tag != dwarf.TagVariable && entry.Tag != dwarf.TagStructType {
			fmt.Printf("not TagVariable\n")
			continue
		}

		var typeOffset dwarf.Offset
		var ok bool
		switch entry.Tag {
		case dwarf.TagVariable:
			typeOffset, ok = entry.Val(dwarf.AttrType).(dwarf.Offset)
			if !ok {
				fmt.Printf("no dwarf.Offset\n")
				continue
			}
		case dwarf.TagStructType:
			typeOffset = entry.Offset
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
			fmt.Printf("<<<<typedefType: %+v\n", typedefType)
			structType = typedefType.Type.(*dwarf.StructType)
			fmt.Printf(">>>>structType: %+v\n", structType)
		}

		pathPrefix := name
		members := collectStructMembers(dwarfData, structType, pathPrefix)
		results = append(results, members...)
	}

	for _, info := range results {
		fmt.Printf("Member: %s, Type: %s, Offset: %d, Size: %d\n", info.Path, info.Type, info.Offset, info.Size)
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
			Type:   field.Type,
		}
		members = append(members, member)

		if subStruct, ok := field.Type.(*dwarf.StructType); ok {
			subMembers := collectStructMembers(dwarfData, subStruct, fieldPath)
			members = append(members, subMembers...)
		}
	}
	return members
}
