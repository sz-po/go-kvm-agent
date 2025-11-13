package display_source

type Commands struct {
	GetDisplayMode        GetDisplayMode        `cmd:"true" help:"Fetch display mode for a display source."`
	GetDisplayPixelFormat GetDisplayPixelFormat `cmd:"true" help:"Fetch display pixel format for a display source."`
	GetDisplayFrameBuffer GetDisplayFrameBuffer `cmd:"true" help:"Fetch display frame buffer for a display source."`
}
