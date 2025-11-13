package display_sink

type Commands struct {
	SetDisplayFrameBufferProvider   SetDisplayFrameBufferProvider   `cmd:"true" help:"Set display frame buffer provider for a display sink."`
	ClearDisplayFrameBufferProvider ClearDisplayFrameBufferProvider `cmd:"true" help:"Clear display frame buffer provider for a display sink."`
}
