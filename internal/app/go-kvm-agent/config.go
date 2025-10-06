package go_kvm_agent

type Config struct {
	Log LogConfig `kong:"embed,prefix='log-'"`
}

type LogConfig struct {
	Level  string `help:"Log level" enum:"debug,info" default:"info"`
	Format string `help:"Log format" enum:"text,json" default:"text"`
}
