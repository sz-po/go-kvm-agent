package main

/*
This file contains all CLI parameters.
*/

// LogParameters defines logging-related CLI parameters.
type LogParameters struct {
	Level  string `help:"Log level" enum:"debug,info" default:"info"`
	Format string `help:"Log format" enum:"text,json" default:"text"`
}

type MachineParameters struct {
	ConfigPath *string `kong:"file" default:"~/.go-kvm-agent/machines/*.yaml"`
}

type Parameters struct {
	ConfigPath string            `kong:"file"`
	Machine    MachineParameters `kong:"embed,prefix='machine-'"`
	Log        LogParameters     `kong:"embed,prefix='log-'"`
}
