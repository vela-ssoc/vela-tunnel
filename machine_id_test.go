package tunnel

import (
	"fmt"
	"log/slog"
	"testing"
)

func TestIdent(t *testing.T) {
	log := slog.Default()
	ident := NewMachineID("", &outLog{log: log})
	for i := 0; i < 1; i++ {
		id := ident.MachineID(true)
		t.Log("machineID:", id)
	}
}

type outLog struct {
	log *slog.Logger
}

func (o *outLog) Infof(s string, a ...any) {
	o.log.Info(fmt.Sprintf(s, a...))
}

func (o *outLog) Warnf(s string, a ...any) {
	o.log.Warn(fmt.Sprintf(s, a...))
}
