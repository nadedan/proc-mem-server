package procio

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type Mmap struct {
	pid       uint64
	mmap      []byte
	isOpen    bool
	isMapped  bool
	fdProcMem int
	addrs     procMapAddrs
	size      uint64
}

func NewMmap(binName string) (*Mmap, error) {
	m := &Mmap{}

	output, err := exec.Command("pidof", binName).Output()
	if err != nil {
		return nil, fmt.Errorf("could not get the pid of %s, perhaps it is not running: %w", binName, err)
	}
	pidStr := string(output)
	pidStr = strings.TrimRight(pidStr, "\n")

	pid, err := strconv.ParseUint(pidStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("pidof gave us '%s', and that cannot be parsed as an int: %w", pidStr, err)
	}
	m.pid = pid

	mapPath := fmt.Sprintf("/proc/%d/maps", m.pid)
	fProcMap, err := os.Open(mapPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %w", mapPath, err)
	}
	defer fProcMap.Close()

	m.addrs, err = readProcMap(fProcMap, binName)
	if err != nil {
		return nil, fmt.Errorf("NewMmap: could not read proc map: %w", err)
	}
	m.size = m.addrs.rwEnd - m.addrs.rwStart

	memPath := fmt.Sprintf("/proc/%d/mem", m.pid)
	m.fdProcMem, err = syscall.Open(memPath, syscall.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("could not open process memory file: %w", err)
	}
	m.isOpen = true

	m.mmap, err = syscall.Mmap(
		m.fdProcMem,
		int64(m.addrs.rwStart),
		int(m.addrs.rwEnd-m.addrs.rwStart),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return nil, fmt.Errorf("could not mmap %s: %w", memPath, err)
	}
	m.isMapped = true

	return m, nil
}

type procMapAddrs struct {
	// base address of the process memory map
	// elf variable virtual addresses are relative to this address
	base uint64
	// start address of the RW region that will be mmapped
	rwStart uint64
	// end address of the RW region that will be mmapped
	rwEnd uint64
}

func readProcMap(f *os.File, binName string) (procMapAddrs, error) {
	m := procMapAddrs{base: 0xFFFF_FFFF_FFFF_FFFF}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, binName) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		addrRange := fields[0]
		addrStrs := strings.Split(addrRange, "-")
		addrStart, err := strconv.ParseUint(addrStrs[0], 16, 64)
		if err != nil {
			return procMapAddrs{}, fmt.Errorf("readProcMap: could not get addrStart: %w", err)
		}
		if addrStart < m.base {
			m.base = addrStart
		}

		perms := fields[1]
		if perms != "rw-p" {
			continue
		}
		addrEnd, err := strconv.ParseUint(addrStrs[1], 16, 64)
		if err != nil {
			return procMapAddrs{}, fmt.Errorf("readProcMap: could not get addrEnd: %w", err)
		}
		m.rwStart = addrStart
		m.rwEnd = addrEnd
	}

	return m, nil
}

func (m *Mmap) Close() {
	if m.isMapped {
		syscall.Munmap(m.mmap)
		m.isMapped = false
	}
	if m.isOpen {
		syscall.Close(m.fdProcMem)
		m.isOpen = false
	}
}

func (m *Mmap) Bytes(offset uint, size uint) []byte {
	offset -= uint(m.addrs.base)
	offset -= uint(m.addrs.rwStart)
	if size > uint(m.size) {
		panic(fmt.Sprintf("cannot get %d bytes from an mmap of size %d", size, m.size))
	}
	return m.mmap[offset : offset+size]
}

func (m *Mmap) Int32(offset uint) int32 {
	b := m.Bytes(offset, 4)
	return int32(binary.LittleEndian.Uint32(b))
}
