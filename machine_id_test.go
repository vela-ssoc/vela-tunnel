package tunnel

import (
	"log/slog"
	"testing"
)

func TestIdent(t *testing.T) {
	log := slog.Default()
	ident := NewMachineID("", log)
	for range 1 {
		id := ident.MachineID(true)
		t.Log("machineID:", id)
	}
}
