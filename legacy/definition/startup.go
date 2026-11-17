package definition

type Startup struct {
	Node    StartupNode      `json:"node"`
	Logger  StartupLogger    `json:"logger"`
	Console StartupConsole   `json:"console"`
	Extends []*StartupExtend `json:"extends"  validate:"omitempty,lte=100,dive"`
}

type StartupNode struct {
	DNS    string `json:"dns"`
	Prefix string `json:"prefix"`
}

type StartupLogger struct {
	Level    string `json:"level"    validate:"oneof=debug info error"`
	Filename string `json:"filename" validate:"required"`
	Console  bool   `json:"console"`
	Format   string `json:"format"   validate:"oneof=text json"`
	Caller   bool   `json:"caller"`
	Skip     int    `json:"skip"     validate:"gt=-20,lt=20"`
}

type StartupConsole struct {
	Enable  bool   `json:"enable"`
	Network string `json:"network" validate:"required_if=Enable true,omitempty,oneof=tcp udp unix"`
	Address string `json:"address" validate:"required_if=Enable true,omitempty,hostname_port"`
	Script  string `json:"script"  validate:"required_if=Enable true"`
}

type StartupExtend struct {
	Name  string `json:"name"  validate:"required"`                                     // 名字
	Type  string `json:"type"  validate:"oneof=number bool string ref string_readonly"` // 类型
	Value string `json:"value" validate:"required"`
}
