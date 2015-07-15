package env

import (
	"fmt"
	"os"
)

type ErrNotSet struct {
	Variable string
}

func (err *ErrNotSet) Error() string {
	return fmt.Sprintf("Environment variable not set: %v", err.Variable)
}

func MustGetenv(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		panic(&ErrNotSet{varName})
	}
	return value
}

func Recover(err *error) {
	if r := recover(); r != nil {
		if ex, ok := r.(*ErrNotSet); ok {
			*err = ex
		} else {
			panic(r)
		}
	}
}
