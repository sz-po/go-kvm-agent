package peripheral

type Commands struct {
	List List `cmd:"true" help:"List peripherals registered on a specific node."`
}
