package mpv

import "fmt"

var (
	NoConfig                = newParameter[Flag]("no-config")
	Cache                   = newParameter[Boolean]("cache")
	Osc                     = newParameter[Boolean]("osc")
	OsdLevel                = newParameter[Integer]("osd-level")
	ForceWindow             = newParameter[ForceWindow_]("force-window")
	Untimed                 = newParameter[Flag]("untimed")
	Demuxer                 = newParameter[Demuxer_]("demuxer")
	NoTerminal              = newParameter[Flag]("no-terminal")
	Title                   = newParameter[String]("title")
	MsgLevel                = newParameter[MsgLevel_]("msg-level")
	OpenGLEarlyFlush        = newParameter[Boolean]("opengl-early-flush")
	SwapChainDepth          = newParameter[Integer]("swapchain-depth")
	DemuxerRawVideoWidth    = newParameter[Integer]("demuxer-rawvideo-w")
	DemuxerRawVideoHeight   = newParameter[Integer]("demuxer-rawvideo-h")
	DemuxerRawVideoFps      = newParameter[Integer]("demuxer-rawvideo-fps")
	DemuxerRawVideoMpFormat = newParameter[RawVideoFormat_]("demuxer-rawvideo-mp-format")
)

type ParameterKey string

type ParameterValue interface {
	String() string
}

type RenderedParameter struct {
	key      ParameterKey
	rendered string
}

type Parameter interface {
	GetKey() ParameterKey
}

type TypedParameter[T ParameterValue] struct {
	Parameter
	key ParameterKey
}

func newParameter[T ParameterValue](key ParameterKey) TypedParameter[T] {
	parameter := TypedParameter[T]{
		key: key,
	}

	return parameter
}

func (parameter TypedParameter[T]) Render(value T) RenderedParameter {
	var rendered string

	if len(value.String()) == 0 {
		rendered = fmt.Sprintf("--%s", parameter.key)
	} else {
		rendered = fmt.Sprintf("--%s=%s", parameter.key, value.String())
	}

	return RenderedParameter{
		key:      parameter.key,
		rendered: rendered,
	}
}

func (parameter TypedParameter[T]) GetKey() ParameterKey {
	return parameter.key
}
