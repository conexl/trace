package collectors

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var procRoot = "/proc"

type socketEntry struct {
	protocol string
	address  string
	port     uint16
	inode    string
}

func collectListeningPorts(ctx context.Context) []ListeningPort {
	entries := append(parseProcNet(filepath.Join(procRoot, "net/tcp"), "tcp4"), parseProcNet(filepath.Join(procRoot, "net/tcp6"), "tcp6")...)
	owners := socketOwners(ctx, procRoot)
	ports := make([]ListeningPort, 0, len(entries))
	for _, entry := range entries {
		port := ListeningPort{Protocol: entry.protocol, Address: entry.address, Port: entry.port}
		if owner, ok := owners[entry.inode]; ok {
			port.PID = owner.pid
			port.Process = owner.name
		}
		ports = append(ports, port)
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Port == ports[j].Port {
			return ports[i].Address < ports[j].Address
		}
		return ports[i].Port < ports[j].Port
	})
	return ports
}

func parseProcNet(path string, protocol string) []socketEntry {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var entries []socketEntry
	scanner := bufio.NewScanner(file)
	first := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if first {
			first = false
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 10 || fields[3] != "0A" {
			continue
		}
		address, port, err := decodeProcAddress(fields[1], protocol)
		if err != nil {
			continue
		}
		entries = append(entries, socketEntry{protocol: protocol, address: address, port: port, inode: fields[9]})
	}
	return entries
}

func decodeProcAddress(value string, protocol string) (string, uint16, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid address %q", value)
	}
	port64, err := strconv.ParseUint(parts[1], 16, 16)
	if err != nil {
		return "", 0, err
	}
	addrHex := parts[0]
	if protocol == "tcp4" {
		bytes, err := hex.DecodeString(addrHex)
		if err != nil || len(bytes) != 4 {
			return "", 0, fmt.Errorf("invalid ipv4 %q", addrHex)
		}
		return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0]).String(), uint16(port64), nil
	}
	bytes, err := hex.DecodeString(addrHex)
	if err != nil || len(bytes) != 16 {
		return "", 0, fmt.Errorf("invalid ipv6 %q", addrHex)
	}
	for i := 0; i < len(bytes); i += 4 {
		bytes[i], bytes[i+3] = bytes[i+3], bytes[i]
		bytes[i+1], bytes[i+2] = bytes[i+2], bytes[i+1]
	}
	return net.IP(bytes).String(), uint16(port64), nil
}

type socketOwner struct {
	pid  int
	name string
}

func socketOwners(ctx context.Context, root string) map[string]socketOwner {
	owners := make(map[string]socketOwner)
	pids, err := filepath.Glob(filepath.Join(root, "[0-9]*"))
	if err != nil {
		return owners
	}
	for _, pidPath := range pids {
		select {
		case <-ctx.Done():
			return owners
		default:
		}
		pid, err := strconv.Atoi(filepath.Base(pidPath))
		if err != nil {
			continue
		}
		name := readProcessName(pidPath)
		fds, _ := filepath.Glob(filepath.Join(pidPath, "fd/*"))
		for _, fd := range fds {
			target, err := os.Readlink(fd)
			if err != nil || !strings.HasPrefix(target, "socket:[") || !strings.HasSuffix(target, "]") {
				continue
			}
			inode := strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]")
			owners[inode] = socketOwner{pid: pid, name: name}
		}
	}
	return owners
}

func readProcessName(pidPath string) string {
	data, err := os.ReadFile(filepath.Join(pidPath, "comm"))
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}
