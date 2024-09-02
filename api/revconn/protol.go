package revconn

type Version byte

const (
	V1 Version = 1
)

type AuthMethod byte

const (
	NoAuth                 AuthMethod = 0
	NoAcceptableAuthMethod AuthMethod = 255
)

const (
	connectionIDLenght   = 8
	connectionInitLength = 2
)
