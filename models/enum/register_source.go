// Path: ./blogX_server/models/enum/register_source.go

package enum

type RegisterSourceType uint8

const (
	RegisterSourceEmailType    RegisterSourceType = 1
	RegisterSourceQQType       RegisterSourceType = 2
	RegisterSourceTerminalType RegisterSourceType = 3
)
