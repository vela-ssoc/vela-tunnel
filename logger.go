package tunnel

import "log"

type Logger interface {
	Infof(string, ...any)
}

type stdLog struct{}

func (sl *stdLog) Infof(s string, a ...any) {
	log.Printf(s, a...)
}
