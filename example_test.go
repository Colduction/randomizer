package randomizer_test

import (
	"fmt"
	"net"

	"github.com/colduction/randomizer-go"
)

func ExampleNetwork_addr() {
	addr := randomizer.Network.Addr(randomizer.IPv6Public)
	prefix := randomizer.Network.Prefix(randomizer.IPv6CIDR, 56)
	host := randomizer.Network.AddrIn(prefix)
	fmt.Println(addr.Is6(), prefix.Contains(host))
	// Output: true true
}

func ExampleNetwork_hardware() {
	static := randomizer.Network.Hardware(nil, randomizer.HardwareBluetoothStatic, randomizer.HardwareOptions{})
	aai := randomizer.Network.Hardware(nil, randomizer.HardwareMACAAI, randomizer.HardwareOptions{})
	eli := randomizer.Network.Hardware(nil, randomizer.HardwareMACELI, randomizer.HardwareOptions{CID: [3]byte{0x0a, 0x1b, 0x2c}})
	fmt.Println(static[0]&0xc0 == 0xc0, aai[0]&0x0f == 0x02, net.HardwareAddr(eli[:3]))
	// Output: true true 0a:1b:2c
}

func ExampleNetwork_iPv6EUI64() {
	mac := net.HardwareAddr{0x00, 0x1a, 0x2b, 0x3c, 0x4d, 0x5e}
	ip := randomizer.Network.IPv6EUI64(nil, randomizer.IPv6LinkLocal, mac)
	fmt.Println(ip)
	// Output: fe80::21a:2bff:fe3c:4d5e
}

func ExampleTelecom_string() {
	imei := randomizer.Telecom.String(randomizer.IMEI, randomizer.TelecomOptions{TAC: "35123456"})
	iccid := randomizer.Telecom.String(randomizer.ICCID, randomizer.TelecomOptions{IIN: "8901", Length: 20})
	fmt.Println(len(imei), imei[:8], len(iccid), iccid[:4])
	// Output: 15 35123456 20 8901
}

func ExampleWord_custom() {
	// Symbols come from the dictionary; adjacent bytes never repeat.
	s := randomizer.Word.Custom("ABC", 6)
	fmt.Println(len(s))
	// Output: 6
}
