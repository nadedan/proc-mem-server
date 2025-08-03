package procio

type sections map[string]section
type section struct {
	addrBase        uint64
	offsetDataStart uint64
	offsetDataEnd   uint64
}

// Addr returns the address where this section starts
func (s section) Addr() uint64 {
	return s.addrBase
}

// AddrDataStart tells us where the required data in this section starts
func (s section) AddrDataStart() uint64 {
	return s.addrBase + s.offsetDataStart
}

// Size returns the number of bytes needed from this section
func (s section) Size() uint64 {
	return s.offsetDataEnd - s.offsetDataStart
}

// RequiredSections looks through the provided global variables
// and determines what sections on the process memory support
// those variables
func RequiredSections(globals variables) sections {
	s := sections{}

	for _, global := range globals {
		thisSection, exists := s[global.Section]
		if !exists {
			thisSection = section{
				addrBase:        global.SectionAddr,
				offsetDataStart: global.Address,
			}
		}

		if global.Address < thisSection.offsetDataStart {
			thisSection.offsetDataStart = global.Address
		}

		endAddr := global.Address + global.Size
		if endAddr > thisSection.offsetDataEnd {
			thisSection.offsetDataEnd = endAddr
		}

		s[global.Section] = thisSection
	}

	return s
}
