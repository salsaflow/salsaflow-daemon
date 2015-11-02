package modules

import "fmt"

type ErrUnknownModuleId struct {
	id string
}

func (err *ErrUnknownModuleId) Error() string {
	return fmt.Sprintf("unknown module id: %v", err.id)
}
