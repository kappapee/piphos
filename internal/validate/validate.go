package validate

import (
	"fmt"
	"net"
)

func Command(nonFlagArgs int) error {
	if nonFlagArgs > 0 {
		return fmt.Errorf("%d unexpected arguments", nonFlagArgs)
	}
	return nil
}

func IP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format for IP %s", ip)
	}
	return nil
}

func Token(token string) error {
	if token == "" {
		return fmt.Errorf("invalid token: token is empty")
	}
	return nil
}
