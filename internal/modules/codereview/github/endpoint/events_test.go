package endpoint

import (
	// Stdlib
	"testing"
)

func Test_eventHandler_interfaces(t *testing.T) {
	if err := ensureInterfaces(); err != nil {
		t.Error(err)
	}
}
