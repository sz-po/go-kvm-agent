//go:build linux

package io

/*
#cgo linux pkg-config: libdrm
#include <drm/drm.h>
#include <drm/drm_mode.h>
#include <drm/drm_fourcc.h>
*/
import "C"

// Basic object/handle identifiers used by DRM legacy (non-atomic) ioctls.
type (
	ObjectHandle   uint32
	ConnectorID    ObjectHandle
	EncoderID      ObjectHandle
	CrtcID         ObjectHandle
	FramebufferID  ObjectHandle
	GemHandle      uint32
	PropertyID     uint32
	ConnectorIndex uint32
)

const (
	InvalidConnectorID   ConnectorID   = 0
	InvalidEncoderID     EncoderID     = 0
	InvalidCrtcID        CrtcID        = 0
	InvalidFramebufferID FramebufferID = 0
	InvalidGemHandle     GemHandle     = 0
)

// ConnectorType mirrors DRM_MODE_CONNECTOR_*.
type ConnectorType uint32

const (
	ConnectorTypeUnknown     ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_Unknown)
	ConnectorTypeVGA         ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_VGA)
	ConnectorTypeDVII        ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_DVII)
	ConnectorTypeDVID        ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_DVID)
	ConnectorTypeDVIA        ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_DVIA)
	ConnectorTypeComposite   ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_Composite)
	ConnectorTypeSVideo      ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_SVIDEO)
	ConnectorTypeLVDS        ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_LVDS)
	ConnectorTypeComponent   ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_Component)
	ConnectorTypeDIN         ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_9PinDIN)
	ConnectorTypeDisplayPort ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_DisplayPort)
	ConnectorTypeHDMIA       ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_HDMIA)
	ConnectorTypeHDMIB       ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_HDMIB)
	ConnectorTypeTV          ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_TV)
	ConnectorTypeEDP         ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_eDP)
	ConnectorTypeVirtual     ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_VIRTUAL)
	ConnectorTypeDSI         ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_DSI)
	ConnectorTypeDPI         ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_DPI)
	ConnectorTypeWriteback   ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_WRITEBACK)
	ConnectorTypeSPI         ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_SPI)
	ConnectorTypeUSB         ConnectorType = ConnectorType(C.DRM_MODE_CONNECTOR_USB)
)

// ConnectionStatus mirrors DRM_MODE_*CONNECTION enums.
type ConnectionStatus uint32

const (
	ConnectionStatusUnknown      ConnectionStatus = ConnectionStatus(C.DRM_MODE_UNKNOWNCONNECTION)
	ConnectionStatusConnected    ConnectionStatus = ConnectionStatus(C.DRM_MODE_CONNECTED)
	ConnectionStatusDisconnected ConnectionStatus = ConnectionStatus(C.DRM_MODE_DISCONNECTED)
)

// EncoderType mirrors DRM_MODE_ENCODER_*.
type EncoderType uint32

const (
	EncoderTypeNone    EncoderType = EncoderType(C.DRM_MODE_ENCODER_NONE)
	EncoderTypeDAC     EncoderType = EncoderType(C.DRM_MODE_ENCODER_DAC)
	EncoderTypeTMDS    EncoderType = EncoderType(C.DRM_MODE_ENCODER_TMDS)
	EncoderTypeLVDS    EncoderType = EncoderType(C.DRM_MODE_ENCODER_LVDS)
	EncoderTypeTVDAC   EncoderType = EncoderType(C.DRM_MODE_ENCODER_TVDAC)
	EncoderTypeVirtual EncoderType = EncoderType(C.DRM_MODE_ENCODER_VIRTUAL)
	EncoderTypeDSI     EncoderType = EncoderType(C.DRM_MODE_ENCODER_DSI)
	EncoderTypeDPMST   EncoderType = EncoderType(C.DRM_MODE_ENCODER_DPMST)
	EncoderTypeDPI     EncoderType = EncoderType(C.DRM_MODE_ENCODER_DPI)
)

// PixelFormat maps to DRM FourCC values.
type PixelFormat uint32

const (
	PixelFormatInvalid     PixelFormat = 0
	PixelFormatXRGB8888    PixelFormat = PixelFormat(C.DRM_FORMAT_XRGB8888)
	PixelFormatARGB8888    PixelFormat = PixelFormat(C.DRM_FORMAT_ARGB8888)
	PixelFormatXBGR8888    PixelFormat = PixelFormat(C.DRM_FORMAT_XBGR8888)
	PixelFormatABGR8888    PixelFormat = PixelFormat(C.DRM_FORMAT_ABGR8888)
	PixelFormatRGB565      PixelFormat = PixelFormat(C.DRM_FORMAT_RGB565)
	PixelFormatNV12        PixelFormat = PixelFormat(C.DRM_FORMAT_NV12)
	PixelFormatNV16        PixelFormat = PixelFormat(C.DRM_FORMAT_NV16)
	PixelFormatNV24        PixelFormat = PixelFormat(C.DRM_FORMAT_NV24)
	PixelFormatYUYV        PixelFormat = PixelFormat(C.DRM_FORMAT_YUYV)
	PixelFormatUYVY        PixelFormat = PixelFormat(C.DRM_FORMAT_UYVY)
	PixelFormatXRGB2101010 PixelFormat = PixelFormat(C.DRM_FORMAT_XRGB2101010)
)

// FramebufferFlag mirrors DRM_MODE_FB_* flags.
type FramebufferFlag uint32

const (
	FramebufferFlagNone      FramebufferFlag = 0
	FramebufferFlagModifiers FramebufferFlag = FramebufferFlag(C.DRM_MODE_FB_MODIFIERS)
)

// PageFlipFlag mirrors DRM_MODE_PAGE_FLIP_*.
type PageFlipFlag uint32

const (
	PageFlipFlagNone            PageFlipFlag = 0
	PageFlipFlagEvent           PageFlipFlag = PageFlipFlag(C.DRM_MODE_PAGE_FLIP_EVENT)
	PageFlipFlagAsync           PageFlipFlag = PageFlipFlag(C.DRM_MODE_PAGE_FLIP_ASYNC)
	PageFlipFlagTargetAbsolute  PageFlipFlag = PageFlipFlag(C.DRM_MODE_PAGE_FLIP_TARGET_ABSOLUTE)
	PageFlipFlagTargetRelative  PageFlipFlag = PageFlipFlag(C.DRM_MODE_PAGE_FLIP_TARGET_RELATIVE)
	PageFlipFlagTargetMonotonic PageFlipFlag = PageFlipFlag(C.DRM_MODE_PAGE_FLIP_TARGET_MONOTONIC)
	PageFlipFlagAllowModeset    PageFlipFlag = PageFlipFlag(C.DRM_MODE_PAGE_FLIP_ALLOW_MODESET)
)

// EventType mirrors DRM_EVENT_*.
type EventType uint32

const (
	EventTypeVBlank           EventType = EventType(C.DRM_EVENT_VBLANK)
	EventTypePageFlipComplete EventType = EventType(C.DRM_EVENT_FLIP_COMPLETE)
)

// ModeInfo matches drm_mode_modeinfo.
type ModeInfo struct {
	Clock      uint32 `json:"clock"`
	HDisplay   uint32 `json:"hDisplay"`
	HSyncStart uint32 `json:"hSyncStart"`
	HSyncEnd   uint32 `json:"hSyncEnd"`
	HTotal     uint32 `json:"hTotal"`
	HSkew      uint32 `json:"hSkew"`

	VDisplay   uint32 `json:"vDisplay"`
	VSyncStart uint32 `json:"vSyncStart"`
	VSyncEnd   uint32 `json:"vSyncEnd"`
	VTotal     uint32 `json:"vTotal"`
	VScan      uint32 `json:"vScan"`

	VRefresh uint32 `json:"vRefresh"`
	Flags    uint32 `json:"flags"`
	Type     uint32 `json:"type"`
	Name     string `json:"name"`
}

var EmptyModeInfo = ModeInfo{}

// CardResources holds IDs exposed by DRM_IOCTL_MODE_GETRESOURCES.
type CardResources struct {
	FramebufferIDs []FramebufferID `json:"framebufferIds"`
	CrtcIDs        []CrtcID        `json:"crtcIds"`
	ConnectorIDs   []ConnectorID   `json:"connectorIds"`
	EncoderIDs     []EncoderID     `json:"encoderIds"`

	MinWidth  uint32 `json:"minWidth"`
	MaxWidth  uint32 `json:"maxWidth"`
	MinHeight uint32 `json:"minHeight"`
	MaxHeight uint32 `json:"maxHeight"`
}

var EmptyCardResources = CardResources{}

// ObjectProperty represents a DRM property/value pair.
type ObjectProperty struct {
	ID    PropertyID `json:"id"`
	Value uint64     `json:"value"`
}

var EmptyObjectProperty = ObjectProperty{}

// Connector mirrors drm_mode_get_connector.
type Connector struct {
	ID         ConnectorID      `json:"id"`
	Index      ConnectorIndex   `json:"index"`
	Type       ConnectorType    `json:"type"`
	Status     ConnectionStatus `json:"status"`
	Encoders   []EncoderID      `json:"encoderIds"`
	Modes      []ModeInfo       `json:"modes"`
	Properties []ObjectProperty `json:"properties"`
}

var EmptyConnector = Connector{}

// Encoder mirrors drm_mode_get_encoder.
type Encoder struct {
	ID             EncoderID   `json:"id"`
	Type           EncoderType `json:"type"`
	CrtcID         CrtcID      `json:"crtcId"`
	PossibleCrtcs  uint32      `json:"possibleCrtcs"`
	PossibleClones uint32      `json:"possibleClones"`
}

var EmptyEncoder = Encoder{}

// Crtc mirrors drm_mode_crtc.
type Crtc struct {
	ID        CrtcID        `json:"id"`
	BufferID  FramebufferID `json:"bufferId"`
	X         uint32        `json:"x"`
	Y         uint32        `json:"y"`
	Width     uint32        `json:"width"`
	Height    uint32        `json:"height"`
	Mode      ModeInfo      `json:"mode"`
	ModeValid bool          `json:"modeValid"`
	GammaSize uint32        `json:"gammaSize"`
}

var EmptyCrtc = Crtc{}

// CrtcConfig is used with DRM_IOCTL_MODE_SETCRTC.
type CrtcConfig struct {
	CrtcID       CrtcID        `json:"crtcId"`
	Framebuffer  FramebufferID `json:"framebufferId"`
	X            uint32        `json:"x"`
	Y            uint32        `json:"y"`
	Mode         ModeInfo      `json:"mode"`
	HasMode      bool          `json:"hasMode"`
	ConnectorIDs []ConnectorID `json:"connectorIds"`
}

var EmptyCrtcConfig = CrtcConfig{}

// DumbBufferSpec wraps dimensions for CREATE_DUMB.
type DumbBufferSpec struct {
	Width        uint32 `json:"width"`
	Height       uint32 `json:"height"`
	BitsPerPixel uint32 `json:"bitsPerPixel"`
}

var EmptyDumbBufferSpec = DumbBufferSpec{}

// DumbBuffer describes a GEM-backed dumb buffer.
type DumbBuffer struct {
	Handle GemHandle      `json:"handle"`
	Pitch  uint32         `json:"pitch"`
	Size   uint32         `json:"size"`
	Spec   DumbBufferSpec `json:"spec"`
}

var EmptyDumbBuffer = DumbBuffer{}

// MappedDumbBuffer includes an mmap'ed view of a dumb buffer.
type MappedDumbBuffer struct {
	DumbBuffer
	Offset uint64 `json:"offset"`
	Data   []byte `json:"-"`
}

var EmptyMappedDumbBuffer = MappedDumbBuffer{}

// FramebufferConfig wraps the data required by DRM_IOCTL_MODE_ADDFB2.
type FramebufferConfig struct {
	Width       uint32          `json:"width"`
	Height      uint32          `json:"height"`
	PixelFormat PixelFormat     `json:"pixelFormat"`
	Handles     [4]GemHandle    `json:"handles"`
	Pitches     [4]uint32       `json:"pitches"`
	Offsets     [4]uint32       `json:"offsets"`
	Flags       FramebufferFlag `json:"flags"`
}

var EmptyFramebufferConfig = FramebufferConfig{}

// PageFlipRequest mirrors drm_mode_crtc_page_flip.
type PageFlipRequest struct {
	CrtcID      CrtcID        `json:"crtcId"`
	Framebuffer FramebufferID `json:"framebufferId"`
	Flags       PageFlipFlag  `json:"flags"`
	UserData    uint64        `json:"userData"`
}

var EmptyPageFlipRequest = PageFlipRequest{}

// Event contains the parsed DRM event header + payload metadata.
type Event struct {
	Type      EventType `json:"type"`
	Sequence  uint32    `json:"sequence"`
	Timestamp int64     `json:"timestamp"`
	UserData  uint64    `json:"userData"`
}

var EmptyEvent = Event{}
