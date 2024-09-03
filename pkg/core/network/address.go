package network

import (
	"fmt"
	"net"
	"strings"
)

var ErrInvalidAddress = fmt.Errorf("invalid address")

const (
	addressSeparator = "."
	addressParts     = 2
)

type Address struct {
	Service   string
	NameSpace string
}

// NewAddress creates a new Address object with the specified service and namespace.
// It returns a pointer to the created Address object.
func NewAddress(service, nameSpace string) *Address {
	return &Address{
		Service:   service,
		NameSpace: nameSpace,
	}
}

// ParseAddress parses the given address string and returns a pointer to an Address object and an error.
// The address string should be in the format "service.namespace", where service and namespace are separated by a dot.
// If the address string is not in the correct format, an error is returned.
// The returned Address object contains the parsed service and namespace values.
func ParseAddress(addr string) (*Address, error) {
	fullAddr, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, ErrInvalidAddress
	}

	addrParts := strings.Split(fullAddr, addressSeparator)
	if len(addrParts) != addressParts {
		return nil, ErrInvalidAddress
	}

	service, nameSpace := addrParts[0], addrParts[1]

	return NewAddress(service, nameSpace), nil
}

// String returns the string representation of the Address object in the format "service.namespace".
func (a *Address) String() string {
	return fmt.Sprintf("%s.%s", a.Service, a.NameSpace)
}
