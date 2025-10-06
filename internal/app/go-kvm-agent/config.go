package go_kvm_agent

// Config contains global configuration for the agent runtime.
type Config struct {
	Log     LogConfig     `kong:"embed,prefix='log-'" json:"log"`
	Machine MachineConfig `kong:"embed,prefix='machine-'" json:"machine"`
}

type MachineConfig struct {
	ConfigPath string `kong:"required" default:"~/.go-kvm-agent/machines/*.yaml" json:"configPath"`
}

// LogConfig defines logging-related configuration options.
type LogConfig struct {
	Level  string `help:"Log level" enum:"debug,info" default:"info" json:"level"`
	Format string `help:"Log format" enum:"text,json" default:"text" json:"format"`
}
