package go_kvm_agent

type Config struct {
	LogLevel  string `help:"Log level" enum:"debug,info" default:"info"`
	LogFormat string `help:"Log format" enum:"text,json" default:"text"`
}
