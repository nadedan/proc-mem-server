package procio

import (
	"fmt"
	"syscall"
)

type Mmap struct {
	pid            uint
	mmap           []byte
	isOpen         bool
	isMapped       bool
	fileDescriptor int
	offset         int64
	size           int
}

func NewMmap(pid uint, offset int64, size int) *Mmap {
	return &Mmap{
		pid:  pid,
		mmap: make([]byte, 0),
	}
}

func (m *Mmap) Open() error {
	memPath := fmt.Sprintf("/proc/%d/mem", m.pid)
	var err error
	m.fileDescriptor, err = syscall.Open(memPath, syscall.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("could not open process memory file: %w", err)
	}
	m.isOpen = true

	m.mmap, err = syscall.Mmap(
		m.fileDescriptor,
		m.offset,
		m.size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return fmt.Errorf("could not mmap %s: %w", memPath, err)
	}
	m.isMapped = true

	return nil
}

func (m *Mmap) Close() {
	if m.isMapped {
		syscall.Munmap(m.mmap)
		m.isMapped = false
	}
	if m.isOpen {
		syscall.Close(m.fileDescriptor)
		m.isOpen = false
	}
}
