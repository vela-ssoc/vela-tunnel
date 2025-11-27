package tunnel

import "testing"

func TestIdent(t *testing.T) {
	ident := NewIdent("", nil)
	for i := 0; i < 10; i++ {
		id := ident.MachineID(true)
		t.Log("machineID:", id)
	}
}
