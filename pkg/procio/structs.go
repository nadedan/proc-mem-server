package procio

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
)

// structFields is a map where the field path is the key and structField info is the value
type structFields map[string]structField
type structField struct {
	// name of the base struct variable
	BaseName string
	// struct field path string from the base to here
	Path string
	// offset in bytes from the base struct var to this field
	Offset uint64
	// size in bytes of the this field's data
	Size uint64
	// data type of this field
	Type dwarf.Type
}

// StructFields looks through an elf.File and produces a map describing all of the struct
// variables within
func StructFields(f *elf.File) (structFields, error) {
	fields := structFields{}

	dwarfData, err := f.DWARF()
	if err != nil {
		return nil, fmt.Errorf("could not get dwarf data from elf file: %w", err)
	}

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

		if entry.Tag != dwarf.TagVariable {
			continue
		}

		typeOffset, ok := entry.Val(dwarf.AttrType).(dwarf.Offset)
		if !ok {
			continue
		}

		typeEntry, err := dwarfData.Type(typeOffset)
		if err != nil {
			fmt.Printf("could not get typEntry\n")
			continue
		}

		structType, ok := asStructType(typeEntry)

		members := collectStructMembers(structType, name, name, 0)
		for _, member := range members {
			fields[member.Path] = member
		}
	}
	return fields, nil
}

// collectStructMembers is a recursive funciton that will traverse a struct
// and document all levels and fields
func collectStructMembers(structType *dwarf.StructType, baseName string, path string, offset uint64) []structField {
	members := []structField{}

	for _, field := range structType.Field {
		if field.Type == nil {
			continue
		}
		newMembers := []structField{}
		structType, isSubStruct := asStructType(field.Type)
		switch isSubStruct {
		case true:
			// drop down into next level of struct
			newMembers = collectStructMembers(
				structType,
				baseName,
				path+"."+field.Name,
				offset+uint64(field.ByteOffset),
			)
		case false:
			newMembers = []structField{
				{
					BaseName: baseName,
					Path:     path + "." + field.Name,
					Offset:   offset + uint64(field.ByteOffset),
					Size:     uint64(field.Type.Size()),
					Type:     field.Type,
				},
			}
		}
		members = append(members, newMembers...)
	}
	return members
}

func asStructType(t dwarf.Type) (*dwarf.StructType, bool) {
	structType, ok := t.(*dwarf.StructType)
	if ok {
		return structType, true
	}
	typedefType, ok := t.(*dwarf.TypedefType)
	if !ok {
		return nil, false
	}

	structType = typedefType.Type.(*dwarf.StructType)

	return structType, true
}
