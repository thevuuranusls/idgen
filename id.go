package idgen

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

const SecondsInDay uint64 = 86400
const NanosecondsInDay = SecondsInDay * 1e9

const (
	BitPrefix       = 20
	BitLenTime      = 26
	BitLenMachineID = 17
)

type Options struct {
	MachineID      func() (uint16, error)
	CheckMachineID func(uint16) bool
}

type Generator struct {
	mutex     *sync.Mutex
	machineID uint16
}

func NewGenerator(opts Options) *Generator {

	var (
		err error
		g   = new(Generator)
	)

	g.mutex = new(sync.Mutex)

	if opts.MachineID == nil {
		g.machineID, err = lower16BitPrivateIP()
	} else {
		g.machineID, err = opts.MachineID()
	}

	if err != nil || (opts.CheckMachineID != nil && !opts.CheckMachineID(g.machineID)) {
		return nil
	}

	return g
}

func (g *Generator) ID() (int, error) {

	g.mutex.Lock()
	defer g.mutex.Unlock()

	prefix := time.Now().Format("060102")
	orgId := uint64(g.machineID)<<(BitLenTime) | elapsedTime()
	return strconv.Atoi(fmt.Sprintf("%s%d", prefix, orgId))
}

// ExtractID returns the machine ID of a ID.
func ExtractID(id uint64) (uint64, uint64) {

	//	TODO: extract and validate prefix string
	//	extract orgId from ID using split by char length
	s := strconv.Itoa(int(id))
	orgID, _ := strconv.Atoi(s[6:])

	//	extract metadata from orgId
	machineID := orgID >> (BitLenTime)
	elapsedTime := orgID & (1<<BitLenTime - 1)
	return uint64(machineID), uint64(elapsedTime)

}

func elapsedTime() uint64 {

	elapsedTime := time.Now().UTC().UnixNano() % int64(NanosecondsInDay) // get nano time in date
	elapsedTime /= 1e6                                                   //  convert to milisecond

	return uint64(elapsedTime)
}

/*
Functions to get a static information for determine ID machine of generator

	privateIPv4: get private IP of machine
	lower16BitPrivateIP: get a lower 2 bytes (16 bits) determine ID default
*/
func privateIPv4() (net.IP, error) {

	// inline function to check ip is private IP version 4
	var isPrivateIPV4 = func(ip net.IP) bool {
		return ip != nil && (ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
	}
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}

		if ip := ipNet.IP.To4(); isPrivateIPV4(ip) {
			return ip, nil
		}
	}

	return nil, errors.New("no private ip address")
}

func lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}
