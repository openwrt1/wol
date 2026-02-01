package magicpacket

import (
	"fmt"
	"net"
)

// MagicPacket represents a wake-on-LAN packet
type MagicPacket struct {
	// The MAC address of the machine to wake up
	MacAddress net.HardwareAddr
}

// NewMagicPacket creates a new MagicPacket for the given MAC address
func NewMagicPacket(macAddress net.HardwareAddr) *MagicPacket {
	return &MagicPacket{MacAddress: macAddress}
}

// Broadcast sends the magic packet to the broadcast address
func (p *MagicPacket) Broadcast() error {
	// Build the actual packet
	packet := make([]byte, 102)
	// Set the synchronization stream (first 6 bytes are 0xFF)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	// Copy the MAC address 16 times into the packet
	for i := 1; i <= 16; i++ {
		copy(packet[i*6:], p.MacAddress)
	}

	// Iterate over all interfaces to send the packet to their broadcast addresses
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	var sent bool
	var lastErr error

	for _, iface := range ifaces {
		// Skip loopback, down, or non-broadcast interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue
			}

			mask := ipNet.Mask
			if len(mask) == net.IPv6len && len(ip4) == net.IPv4len {
				mask = mask[12:]
			}

			// Calculate broadcast address
			broadcastIP := make(net.IP, len(ip4))
			for i := range ip4 {
				broadcastIP[i] = ip4[i] | ^mask[i]
			}

			addr := &net.UDPAddr{
				IP:   broadcastIP,
				Port: 9,
			}

			conn, err := net.DialUDP("udp", nil, addr)
			if err != nil {
				lastErr = err
				continue
			}

			_, err = conn.Write(packet)
			conn.Close()
			if err != nil {
				lastErr = err
				continue
			}
			sent = true
		}
	}

	// If we managed to send to at least one interface, consider it a success.
	// Otherwise, try the global broadcast address as a fallback.
	if !sent {
		addr := &net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: 9,
		}
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			if lastErr != nil {
				return fmt.Errorf("failed to send packet: %v (last error)", lastErr)
			}
			return err
		}
		defer conn.Close()
		_, err = conn.Write(packet)
		return err
	}

	return nil
}

// Send sends the magic packet to a specific address (unicast)
func (p *MagicPacket) Send(addr string) error {
	// Build the actual packet
	packet := make([]byte, 102)
	// Set the synchronization stream (first 6 bytes are 0xFF)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	// Copy the MAC address 16 times into the packet
	for i := 1; i <= 16; i++ {
		copy(packet[i*6:], p.MacAddress)
	}

	conn, err := net.Dial("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}
