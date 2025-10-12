package mpv

import "fmt"

type ForceWindow_ string

const (
	ForceWindowImmediate ForceWindow_ = "immediate"
)

func (value ForceWindow_) String() string {
	return string(value)
}

type Flag bool

func (value Flag) String() string {
	return ""
}

type Integer int

func (value Integer) String() string {
	return fmt.Sprintf("%d", value)
}

type Boolean bool

func (value Boolean) String() string {
	if value {
		return "yes"
	} else {
		return "no"
	}
}

type String string

func (value String) String() string {
	return fmt.Sprintf("\"%s\"", string(value))
}

type Demuxer_ string

const (
	DemuxerRawVideo Demuxer_ = "rawvideo"
)

func (value Demuxer_) String() string {
	return string(value)
}

type MsgLevel_ string

const (
	MsgLevelAllFatal MsgLevel_ = "all=fatal"
	MsgLevelAllError MsgLevel_ = "all=error"
	MsgLevelAllWarn  MsgLevel_ = "all=warn"
	MsgLevelAllInfo  MsgLevel_ = "all=info"
)

func (value MsgLevel_) String() string {
	return string(value)
}

type RawVideoFormat_ string

const (
	RawVideoFormatRGB24 RawVideoFormat_ = "rgb24"
)

func (value RawVideoFormat_) String() string {
	return string(value)
}
