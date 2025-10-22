package edid

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type DetailedTimingsBlock [72]byte

type DetailedTimingsMonitorNameDescriptorEntry struct {
	Name string `json:"name" validate:"required,ascii,min=1,max=13"`
}

type DetailedTimingsMonitorSerialEntry struct {
	Name string `json:"name" validate:"required,ascii,min=1,max=13"`
}

type DetailedTimingsRangeLimitsDescriptorEntry struct {
	MinVerticalHz uint8 `json:"minVerticalHz" validate:"gte=1,lte=240,ltfield=MaxVerticalHz"`
	MaxVerticalHz uint8 `json:"maxVerticalHz" validate:"gtfield=MinVerticalHz,lte=480"`
	MinKHz        uint8 `json:"minKHz"        validate:"gte=1,lte=255,ltfield=MaxKHz"`
	MaxKHz        uint8 `json:"maxKHz"        validate:"gtfield=MinKHz,lte=255"`
	MaxClockMHz   uint8 `json:"maxClockMHz"   validate:"gte=10,lte=255"`
}

type DetailedTimingsStandardDescriptorEntry struct {
	PixelClock           uint32 `json:"pixelClock"            validate:"required,gt=0,lte=655350"`
	HorizontalActive     uint16 `json:"horizontalActive"      validate:"required,gte=16,lte=4095"`
	HorizontalBlank      uint16 `json:"horizontalBlank"       validate:"required,gte=1,lte=4095"`
	VerticalActive       uint16 `json:"verticalActive"        validate:"required,gte=16,lte=4095"`
	VerticalBlank        uint16 `json:"verticalBlank"         validate:"required,gte=1,lte=2047"`
	HorizontalSyncOffset uint16 `json:"horizontalSyncOffset"  validate:"required,gte=0,lte=1023"`
	HorizontalSyncWidth  uint16 `json:"horizontalSyncWidth"   validate:"required,gte=0,lte=1023"`
	VerticalSyncOffset   uint16 `json:"verticalSyncOffset"    validate:"required,gte=0,lte=63"`
	VerticalSyncWidth    uint16 `json:"verticalSyncWidth"     validate:"required,gte=0,lte=63"`
	HorizontalImageSize  uint16 `json:"horizontalImageSIZE" validate:"required,gte=0,lte=10000"`
	VerticalImageSize    uint16 `json:"verticalImageSIZE" validate:"required,gte=0,lte=10000"`
	HorizontalBorder     uint8  `json:"horizontalBorder"      validate:"gte=0,lte=255"`
	VerticalBorder       uint8  `json:"verticalBorder"        validate:"gte=0,lte=255"`

	Interlaced bool `json:"interlaced"`

	StereoMode DetailedTimingsStereoMode `json:"stereoMode"`
	SyncType   DetailedTimingsSyncType   `json:"syncType"`

	HorizontalSyncPositive   bool `json:"horizontalSyncPositive"`
	VerticalSyncPositive     bool `json:"verticalSyncPositive"`
	CompositeSyncSerration   bool `json:"compositeSyncSerration"`
	CompositeSyncOnAllColors bool `json:"compositeSyncOnAllColors"`
	CompositeSyncOnGreenOnly bool `json:"compositeSyncOnGreenOnly"`
	CompositeSyncOnRedBlue   bool `json:"compositeSyncOnRedBlue"`
}

type DetailedTimingsEntrySpecification struct {
	Standard      *DetailedTimingsStandardDescriptorEntry    `json:"standard,omitempty"`
	RangeLimits   *DetailedTimingsRangeLimitsDescriptorEntry `json:"rangeLimits,omitempty"`
	MonitorName   *DetailedTimingsMonitorNameDescriptorEntry `json:"monitorName,omitempty"`
	MonitorSerial *DetailedTimingsMonitorSerialEntry         `json:"monitorSerial,omitempty"`
}

type DetailedTimingsStereoMode uint8

const (
	DetailedTimingsStereoModeUnknown DetailedTimingsStereoMode = iota
	DetailedTimingsStereoModeNone
	DetailedTimingsStereoModeFieldSequentialRight
	DetailedTimingsStereoModeFieldSequentialLeft
	DetailedTimingsStereoModeTwoWayInterleaved
)

type DetailedTimingsSyncType uint8

const (
	DetailedTimingsSyncTypeUnknown DetailedTimingsSyncType = iota
	DetailedTimingsSyncTypeAnalogComposite
	DetailedTimingsSyncTypeAnalogBipolar
	DetailedTimingsSyncTypeDigitalComposite
	DetailedTimingsSyncTypeDigitalSeparate
)

type DetailedTimingsSpecification struct {
	Entries []DetailedTimingsEntrySpecification `json:"entries,omitempty" validate:"omitempty,dive"`
}

const (
	detailedTimingDescriptorLength = 18
	detailedTimingDescriptorCount  = len(DetailedTimingsBlock{}) / detailedTimingDescriptorLength

	detailedTimingDescriptorTextDataLength = 13
	detailedTimingDescriptorTextMaxPayload = detailedTimingDescriptorTextDataLength - 1

	detailedTimingDescriptorTagRangeLimits   = 0xFD
	detailedTimingDescriptorTagMonitorName   = 0xFC
	detailedTimingDescriptorTagMonitorSerial = 0xFF
	detailedTimingDescriptorTagAsciiString   = 0xFE
	detailedTimingDescriptorTagUnused        = 0x10
)

func CreateDetailedTimingsSpecificationFromBlock(block DetailedTimingsBlock) (*DetailedTimingsSpecification, error) {
	specification := &DetailedTimingsSpecification{}

	for descriptorIndex := 0; descriptorIndex < len(block); descriptorIndex += detailedTimingDescriptorLength {
		descriptor := block[descriptorIndex : descriptorIndex+detailedTimingDescriptorLength]

		entry, err := decodeDetailedTimingDescriptor(descriptor)
		if err != nil {
			return nil, fmt.Errorf("decode descriptor %d: %w", descriptorIndex/detailedTimingDescriptorLength, err)
		}

		if err := entry.Validate(); err != nil {
			return nil, fmt.Errorf("validate descriptor %d: %w", descriptorIndex/detailedTimingDescriptorLength, err)
		}

		if !entry.isEmpty() {
			specification.Entries = append(specification.Entries, entry)
		}
	}

	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	return specification, nil
}

func CreateDetailedTimingsBlockFromSpecification(specification DetailedTimingsSpecification) (*DetailedTimingsBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	block := &DetailedTimingsBlock{}

	for index := 0; index < detailedTimingDescriptorCount; index++ {
		var entry DetailedTimingsEntrySpecification
		if index < len(specification.Entries) {
			entry = specification.Entries[index]
		}

		if err := entry.Validate(); err != nil {
			return nil, fmt.Errorf("validate entry %d: %w", index, err)
		}

		descriptor, err := encodeDetailedTimingDescriptor(entry)
		if err != nil {
			return nil, fmt.Errorf("encode descriptor %d: %w", index, err)
		}

		copy(block[index*detailedTimingDescriptorLength:], descriptor[:])
	}

	return block, nil
}

func (specification *DetailedTimingsSpecification) Validate() error {
	if specification == nil {
		return fmt.Errorf("nil specification")
	}

	if len(specification.Entries) > detailedTimingDescriptorCount {
		return fmt.Errorf("too many detailed timing entries: %d", len(specification.Entries))
	}

	for index := range specification.Entries {
		if err := specification.Entries[index].Validate(); err != nil {
			return fmt.Errorf("entry %d invalid: %w", index, err)
		}
	}

	return nil
}

func (specification *DetailedTimingsEntrySpecification) Validate() error {
	if specification == nil {
		return fmt.Errorf("nil entry")
	}

	descriptorCount := 0

	if specification.Standard != nil {
		descriptorCount++
		if err := specification.Standard.Validate(); err != nil {
			return fmt.Errorf("standard descriptor invalid: %w", err)
		}
	}

	if specification.RangeLimits != nil {
		descriptorCount++
		if err := specification.RangeLimits.Validate(); err != nil {
			return fmt.Errorf("range limits descriptor invalid: %w", err)
		}
	}

	if specification.MonitorName != nil {
		descriptorCount++
		if err := specification.MonitorName.Validate(); err != nil {
			return fmt.Errorf("monitor name descriptor invalid: %w", err)
		}
	}

	if specification.MonitorSerial != nil {
		descriptorCount++
		if err := specification.MonitorSerial.Validate(); err != nil {
			return fmt.Errorf("monitor serial descriptor invalid: %w", err)
		}
	}

	if descriptorCount > 1 {
		return fmt.Errorf("multiple descriptor types provided")
	}

	return nil
}

func (specification DetailedTimingsEntrySpecification) isEmpty() bool {
	return specification.Standard == nil &&
		specification.RangeLimits == nil &&
		specification.MonitorName == nil &&
		specification.MonitorSerial == nil
}

func (descriptor *DetailedTimingsStandardDescriptorEntry) Validate() error {
	if descriptor == nil {
		return fmt.Errorf("nil standard descriptor")
	}

	specificationValidator := validator.New(validator.WithRequiredStructEnabled())

	if err := specificationValidator.Struct(descriptor); err != nil {
		return err
	}

	switch descriptor.StereoMode {
	case DetailedTimingsStereoModeNone,
		DetailedTimingsStereoModeFieldSequentialRight,
		DetailedTimingsStereoModeFieldSequentialLeft,
		DetailedTimingsStereoModeTwoWayInterleaved:
	default:
		return fmt.Errorf("unsupported stereo mode: %d", descriptor.StereoMode)
	}

	switch descriptor.SyncType {
	case DetailedTimingsSyncTypeAnalogComposite,
		DetailedTimingsSyncTypeAnalogBipolar,
		DetailedTimingsSyncTypeDigitalComposite,
		DetailedTimingsSyncTypeDigitalSeparate:
	default:
		return fmt.Errorf("unsupported sync type: %d", descriptor.SyncType)
	}

	if descriptor.PixelClock%10 != 0 {
		return fmt.Errorf("pixel clock must be divisible by 10")
	}

	if descriptor.HorizontalSyncOffset+descriptor.HorizontalSyncWidth > descriptor.HorizontalBlank {
		return fmt.Errorf("horizontal sync exceeds blanking interval")
	}

	if descriptor.VerticalSyncOffset+descriptor.VerticalSyncWidth > descriptor.VerticalBlank {
		return fmt.Errorf("vertical sync exceeds blanking interval")
	}

	compositeLocationFlags := []bool{
		descriptor.CompositeSyncOnAllColors,
		descriptor.CompositeSyncOnGreenOnly,
		descriptor.CompositeSyncOnRedBlue,
	}

	locationCount := 0
	for _, flag := range compositeLocationFlags {
		if flag {
			locationCount++
		}
	}

	switch descriptor.SyncType {
	case DetailedTimingsSyncTypeAnalogComposite:
		if locationCount != 1 {
			return fmt.Errorf("analog composite sync must set exactly one location flag")
		}
		if descriptor.HorizontalSyncPositive || descriptor.VerticalSyncPositive {
			return fmt.Errorf("analog composite sync cannot set digital polarity flags")
		}
	case DetailedTimingsSyncTypeAnalogBipolar:
		if locationCount != 0 {
			return fmt.Errorf("analog bipolar sync does not support composite sync location flags")
		}
		if descriptor.CompositeSyncSerration {
			return fmt.Errorf("analog bipolar sync does not support composite sync serration")
		}
		if descriptor.HorizontalSyncPositive || descriptor.VerticalSyncPositive {
			return fmt.Errorf("analog bipolar sync cannot set digital polarity flags")
		}
	case DetailedTimingsSyncTypeDigitalComposite:
		if locationCount != 0 {
			return fmt.Errorf("digital composite sync does not support analog location flags")
		}
		if descriptor.HorizontalSyncPositive || descriptor.VerticalSyncPositive {
			return fmt.Errorf("digital composite sync cannot set separate polarity flags")
		}
	case DetailedTimingsSyncTypeDigitalSeparate:
		if descriptor.CompositeSyncSerration {
			return fmt.Errorf("digital separate sync does not support composite serration")
		}
		if locationCount != 0 {
			return fmt.Errorf("digital separate sync does not support analog location flags")
		}
	default:
		return fmt.Errorf("unknown sync type: %d", descriptor.SyncType)
	}

	if descriptor.SyncType != DetailedTimingsSyncTypeAnalogComposite && descriptor.SyncType != DetailedTimingsSyncTypeDigitalComposite {
		if descriptor.CompositeSyncSerration {
			return fmt.Errorf("composite sync serration only applies to composite sync types")
		}
	}

	return nil
}

func (descriptor *DetailedTimingsRangeLimitsDescriptorEntry) Validate() error {
	if descriptor == nil {
		return fmt.Errorf("nil range limits descriptor")
	}

	specificationValidator := validator.New(validator.WithRequiredStructEnabled())

	if err := specificationValidator.Struct(descriptor); err != nil {
		return err
	}

	return nil
}

func (descriptor *DetailedTimingsMonitorNameDescriptorEntry) Validate() error {
	if descriptor == nil {
		return fmt.Errorf("nil monitor name descriptor")
	}

	specificationValidator := validator.New(validator.WithRequiredStructEnabled())

	if err := specificationValidator.Struct(descriptor); err != nil {
		return err
	}

	if strings.ContainsRune(descriptor.Name, '\n') {
		return fmt.Errorf("monitor name must not contain a newline")
	}

	if len(descriptor.Name) > detailedTimingDescriptorTextMaxPayload {
		return fmt.Errorf("monitor name too long: %d", len(descriptor.Name))
	}

	return nil
}

func (descriptor *DetailedTimingsMonitorSerialEntry) Validate() error {
	if descriptor == nil {
		return fmt.Errorf("nil monitor serial descriptor")
	}

	specificationValidator := validator.New(validator.WithRequiredStructEnabled())

	if err := specificationValidator.Struct(descriptor); err != nil {
		return err
	}

	if strings.ContainsRune(descriptor.Name, '\n') {
		return fmt.Errorf("monitor serial must not contain a newline")
	}

	if len(descriptor.Name) > detailedTimingDescriptorTextMaxPayload {
		return fmt.Errorf("monitor serial too long: %d", len(descriptor.Name))
	}

	return nil
}

func decodeDetailedTimingDescriptor(descriptor []byte) (DetailedTimingsEntrySpecification, error) {
	if len(descriptor) != detailedTimingDescriptorLength {
		return DetailedTimingsEntrySpecification{}, fmt.Errorf("descriptor length mismatch: %d", len(descriptor))
	}

	if isZeroDescriptor(descriptor) {
		return DetailedTimingsEntrySpecification{}, nil
	}

	pixelClock := uint16(descriptor[0]) | uint16(descriptor[1])<<8

	if pixelClock != 0 {
		standard, err := decodeStandardTimingDescriptor(descriptor, pixelClock)
		if err != nil {
			return DetailedTimingsEntrySpecification{}, err
		}

		return DetailedTimingsEntrySpecification{Standard: standard}, nil
	}

	return decodeMonitorDescriptor(descriptor)
}

func decodeStandardTimingDescriptor(descriptor []byte, rawPixelClock uint16) (*DetailedTimingsStandardDescriptorEntry, error) {
	horizontalActive := decode12Bit(descriptor[2], descriptor[4]>>4)
	horizontalBlank := decode12Bit(descriptor[3], descriptor[4]&0x0F)
	verticalActive := decode12Bit(descriptor[5], descriptor[7]>>4)
	verticalBlank := decode12Bit(descriptor[6], descriptor[7]&0x0F)

	horizontalSyncOffset := decode10Bit(descriptor[8], descriptor[11]>>6)
	horizontalSyncWidth := decode10Bit(descriptor[9], (descriptor[11]>>4)&0x03)

	verticalSyncOffset := decode6Bit(descriptor[10]>>4, (descriptor[11]>>2)&0x03)
	verticalSyncWidth := decode6Bit(descriptor[10]&0x0F, descriptor[11]&0x03)

	horizontalImage := decode12Bit(descriptor[12], descriptor[14]>>4)
	verticalImage := decode12Bit(descriptor[13], descriptor[14]&0x0F)

	entry := &DetailedTimingsStandardDescriptorEntry{
		PixelClock:           uint32(rawPixelClock) * 10,
		HorizontalActive:     horizontalActive,
		HorizontalBlank:      horizontalBlank,
		VerticalActive:       verticalActive,
		VerticalBlank:        verticalBlank,
		HorizontalSyncOffset: horizontalSyncOffset,
		HorizontalSyncWidth:  horizontalSyncWidth,
		VerticalSyncOffset:   verticalSyncOffset,
		VerticalSyncWidth:    verticalSyncWidth,
		HorizontalImageSize:  horizontalImage,
		VerticalImageSize:    verticalImage,
		HorizontalBorder:     descriptor[15],
		VerticalBorder:       descriptor[16],
	}

	if err := decodeDetailedTimingFlags(descriptor[17], entry); err != nil {
		return nil, err
	}

	if err := entry.Validate(); err != nil {
		return nil, err
	}

	return entry, nil
}

func decodeDetailedTimingFlags(flagValue byte, entry *DetailedTimingsStandardDescriptorEntry) error {
	if entry == nil {
		return fmt.Errorf("nil standard descriptor")
	}

	entry.Interlaced = flagValue&0x80 != 0

	stereoBits := (flagValue >> 5) & 0x03
	stereoMode, err := decodeStereoMode(stereoBits)
	if err != nil {
		return err
	}
	entry.StereoMode = stereoMode

	syncBits := (flagValue >> 3) & 0x03
	syncType, err := decodeSyncType(syncBits)
	if err != nil {
		return err
	}
	entry.SyncType = syncType

	entry.HorizontalSyncPositive = false
	entry.VerticalSyncPositive = false
	entry.CompositeSyncSerration = false
	entry.CompositeSyncOnAllColors = false
	entry.CompositeSyncOnGreenOnly = false
	entry.CompositeSyncOnRedBlue = false

	detailBits := flagValue & 0x07

	switch syncType {
	case DetailedTimingsSyncTypeAnalogComposite:
		if detailBits&0x04 != 0 {
			entry.CompositeSyncSerration = true
		}

		locationBits := detailBits & 0x03
		switch locationBits {
		case 0x00:
			entry.CompositeSyncOnGreenOnly = true
		case 0x01:
			entry.CompositeSyncOnAllColors = true
		case 0x02:
			entry.CompositeSyncOnRedBlue = true
		default:
			return fmt.Errorf("unsupported analog composite sync location bits: 0x%02x", locationBits)
		}
	case DetailedTimingsSyncTypeAnalogBipolar:
		if detailBits != 0 {
			return fmt.Errorf("unsupported analog bipolar sync detail bits: 0x%02x", detailBits)
		}
	case DetailedTimingsSyncTypeDigitalComposite:
		if detailBits&0x04 != 0 {
			entry.CompositeSyncSerration = true
		}
		if detailBits&0x03 != 0 {
			return fmt.Errorf("unsupported digital composite sync detail bits: 0x%02x", detailBits&0x03)
		}
	case DetailedTimingsSyncTypeDigitalSeparate:
		if detailBits&0x01 != 0 {
			return fmt.Errorf("unsupported digital separate sync reserved bit: 0x%02x", detailBits&0x01)
		}
		if detailBits&0x04 != 0 {
			entry.VerticalSyncPositive = true
		}
		if detailBits&0x02 != 0 {
			entry.HorizontalSyncPositive = true
		}
	default:
		return fmt.Errorf("unsupported sync type: %d", syncType)
	}

	return nil
}

func decodeStereoMode(bits byte) (DetailedTimingsStereoMode, error) {
	switch bits {
	case 0x00:
		return DetailedTimingsStereoModeNone, nil
	case 0x01:
		return DetailedTimingsStereoModeFieldSequentialRight, nil
	case 0x02:
		return DetailedTimingsStereoModeFieldSequentialLeft, nil
	case 0x03:
		return DetailedTimingsStereoModeTwoWayInterleaved, nil
	default:
		return DetailedTimingsStereoModeUnknown, fmt.Errorf("unsupported stereo bits value: 0x%02x", bits)
	}
}

func decodeSyncType(bits byte) (DetailedTimingsSyncType, error) {
	switch bits {
	case 0x00:
		return DetailedTimingsSyncTypeAnalogComposite, nil
	case 0x01:
		return DetailedTimingsSyncTypeAnalogBipolar, nil
	case 0x02:
		return DetailedTimingsSyncTypeDigitalComposite, nil
	case 0x03:
		return DetailedTimingsSyncTypeDigitalSeparate, nil
	default:
		return DetailedTimingsSyncTypeUnknown, fmt.Errorf("unsupported sync bits value: 0x%02x", bits)
	}
}

func decodeMonitorDescriptor(descriptor []byte) (DetailedTimingsEntrySpecification, error) {
	if descriptor[2] != 0 || descriptor[4] != 0 {
		return DetailedTimingsEntrySpecification{}, fmt.Errorf("unexpected monitor descriptor header")
	}

	tag := descriptor[3]

	switch tag {
	case detailedTimingDescriptorTagRangeLimits:
		entry := &DetailedTimingsRangeLimitsDescriptorEntry{
			MinVerticalHz: descriptor[5],
			MaxVerticalHz: descriptor[6],
			MinKHz:        descriptor[7],
			MaxKHz:        descriptor[8],
			MaxClockMHz:   descriptor[9],
		}

		defaultGTFParameters := []byte{0x00, 0x0A, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
		extension := descriptor[10:]
		if !bytes.Equal(extension, make([]byte, len(extension))) && !bytes.Equal(extension, defaultGTFParameters) {
			return DetailedTimingsEntrySpecification{}, fmt.Errorf("unsupported range descriptor extension")
		}

		if err := entry.Validate(); err != nil {
			return DetailedTimingsEntrySpecification{}, err
		}

		return DetailedTimingsEntrySpecification{RangeLimits: entry}, nil
	case detailedTimingDescriptorTagMonitorName:
		name, err := decodeTextDescriptor(descriptor[5:])
		if err != nil {
			return DetailedTimingsEntrySpecification{}, fmt.Errorf("decode monitor name: %w", err)
		}

		entry := &DetailedTimingsMonitorNameDescriptorEntry{Name: name}

		if err := entry.Validate(); err != nil {
			return DetailedTimingsEntrySpecification{}, err
		}

		return DetailedTimingsEntrySpecification{MonitorName: entry}, nil
	case detailedTimingDescriptorTagMonitorSerial:
		serial, err := decodeTextDescriptor(descriptor[5:])
		if err != nil {
			return DetailedTimingsEntrySpecification{}, fmt.Errorf("decode monitor serial: %w", err)
		}

		entry := &DetailedTimingsMonitorSerialEntry{Name: serial}

		if err := entry.Validate(); err != nil {
			return DetailedTimingsEntrySpecification{}, err
		}

		return DetailedTimingsEntrySpecification{MonitorSerial: entry}, nil
	case detailedTimingDescriptorTagUnused:
		if !isBlankDescriptorPayload(descriptor[5:]) {
			return DetailedTimingsEntrySpecification{}, fmt.Errorf("unsupported unused descriptor payload")
		}

		return DetailedTimingsEntrySpecification{}, nil
	case detailedTimingDescriptorTagAsciiString:
		return DetailedTimingsEntrySpecification{}, fmt.Errorf("ascii string descriptors are not supported")
	default:
		return DetailedTimingsEntrySpecification{}, fmt.Errorf("unsupported monitor descriptor tag: 0x%02x", tag)
	}
}

func encodeDetailedTimingDescriptor(entry DetailedTimingsEntrySpecification) ([detailedTimingDescriptorLength]byte, error) {
	if entry.Standard != nil {
		return encodeStandardTimingDescriptor(entry.Standard)
	}

	if entry.RangeLimits != nil {
		return encodeRangeLimitsDescriptor(entry.RangeLimits)
	}

	if entry.MonitorName != nil {
		return encodeMonitorTextDescriptor(detailedTimingDescriptorTagMonitorName, entry.MonitorName.Name)
	}

	if entry.MonitorSerial != nil {
		return encodeMonitorTextDescriptor(detailedTimingDescriptorTagMonitorSerial, entry.MonitorSerial.Name)
	}

	var descriptor [detailedTimingDescriptorLength]byte
	descriptor[3] = detailedTimingDescriptorTagUnused

	return descriptor, nil
}

func encodeStandardTimingDescriptor(entry *DetailedTimingsStandardDescriptorEntry) ([detailedTimingDescriptorLength]byte, error) {
	if entry == nil {
		return [detailedTimingDescriptorLength]byte{}, fmt.Errorf("nil standard descriptor")
	}

	if err := entry.Validate(); err != nil {
		return [detailedTimingDescriptorLength]byte{}, err
	}

	if entry.PixelClock%10 != 0 {
		return [detailedTimingDescriptorLength]byte{}, fmt.Errorf("pixel clock must be divisible by 10")
	}

	pixelClockUnits := uint16(entry.PixelClock / 10)

	var descriptor [detailedTimingDescriptorLength]byte

	descriptor[0] = byte(pixelClockUnits & 0xFF)
	descriptor[1] = byte(pixelClockUnits >> 8)

	descriptor[2] = byte(entry.HorizontalActive & 0xFF)
	descriptor[3] = byte(entry.HorizontalBlank & 0xFF)
	descriptor[4] = byte((((entry.HorizontalActive >> 8) & 0x0F) << 4) | ((entry.HorizontalBlank >> 8) & 0x0F))

	descriptor[5] = byte(entry.VerticalActive & 0xFF)
	descriptor[6] = byte(entry.VerticalBlank & 0xFF)
	descriptor[7] = byte((((entry.VerticalActive >> 8) & 0x0F) << 4) | ((entry.VerticalBlank >> 8) & 0x0F))

	horizontalSyncOffsetLow, horizontalSyncOffsetHigh := split10Bit(entry.HorizontalSyncOffset)
	descriptor[8] = horizontalSyncOffsetLow

	horizontalSyncWidthLow, horizontalSyncWidthHigh := split10Bit(entry.HorizontalSyncWidth)
	descriptor[9] = horizontalSyncWidthLow

	verticalSyncOffsetLow, verticalSyncOffsetHigh := split6Bit(entry.VerticalSyncOffset)
	verticalSyncWidthLow, verticalSyncWidthHigh := split6Bit(entry.VerticalSyncWidth)

	descriptor[10] = byte((verticalSyncOffsetLow << 4) | verticalSyncWidthLow)
	descriptor[11] = byte((horizontalSyncOffsetHigh << 6) | (horizontalSyncWidthHigh << 4) | (verticalSyncOffsetHigh << 2) | verticalSyncWidthHigh)

	descriptor[12] = byte(entry.HorizontalImageSize & 0xFF)
	descriptor[13] = byte(entry.VerticalImageSize & 0xFF)
	descriptor[14] = byte((((entry.HorizontalImageSize >> 8) & 0x0F) << 4) | ((entry.VerticalImageSize >> 8) & 0x0F))

	descriptor[15] = entry.HorizontalBorder
	descriptor[16] = entry.VerticalBorder

	flagValue, err := encodeDetailedTimingFlags(entry)
	if err != nil {
		return [detailedTimingDescriptorLength]byte{}, err
	}

	descriptor[17] = flagValue

	return descriptor, nil
}

func encodeDetailedTimingFlags(entry *DetailedTimingsStandardDescriptorEntry) (byte, error) {
	if entry == nil {
		return 0, fmt.Errorf("nil standard descriptor")
	}

	flagValue := byte(0)
	if entry.Interlaced {
		flagValue |= 0x80
	}

	stereoBits, err := encodeStereoMode(entry.StereoMode)
	if err != nil {
		return 0, err
	}
	flagValue |= stereoBits << 5

	syncBits, err := encodeSyncType(entry.SyncType)
	if err != nil {
		return 0, err
	}
	flagValue |= syncBits << 3

	switch entry.SyncType {
	case DetailedTimingsSyncTypeAnalogComposite:
		if entry.CompositeSyncSerration {
			flagValue |= 0x04
		}

		switch {
		case entry.CompositeSyncOnGreenOnly:
			// 0x00 indicates sync on green only
		case entry.CompositeSyncOnAllColors:
			flagValue |= 0x01
		case entry.CompositeSyncOnRedBlue:
			flagValue |= 0x02
		default:
			return 0, fmt.Errorf("analog composite sync must choose a location flag")
		}
	case DetailedTimingsSyncTypeAnalogBipolar:
		// No additional bits defined
	case DetailedTimingsSyncTypeDigitalComposite:
		if entry.CompositeSyncSerration {
			flagValue |= 0x04
		}
	case DetailedTimingsSyncTypeDigitalSeparate:
		if entry.VerticalSyncPositive {
			flagValue |= 0x04
		}
		if entry.HorizontalSyncPositive {
			flagValue |= 0x02
		}
	default:
		return 0, fmt.Errorf("unsupported sync type: %d", entry.SyncType)
	}

	return flagValue, nil
}

func encodeStereoMode(mode DetailedTimingsStereoMode) (byte, error) {
	switch mode {
	case DetailedTimingsStereoModeNone:
		return 0x00, nil
	case DetailedTimingsStereoModeFieldSequentialRight:
		return 0x01, nil
	case DetailedTimingsStereoModeFieldSequentialLeft:
		return 0x02, nil
	case DetailedTimingsStereoModeTwoWayInterleaved:
		return 0x03, nil
	default:
		return 0, fmt.Errorf("unsupported stereo mode: %d", mode)
	}
}

func encodeSyncType(syncType DetailedTimingsSyncType) (byte, error) {
	switch syncType {
	case DetailedTimingsSyncTypeAnalogComposite:
		return 0x00, nil
	case DetailedTimingsSyncTypeAnalogBipolar:
		return 0x01, nil
	case DetailedTimingsSyncTypeDigitalComposite:
		return 0x02, nil
	case DetailedTimingsSyncTypeDigitalSeparate:
		return 0x03, nil
	default:
		return 0, fmt.Errorf("unsupported sync type: %d", syncType)
	}
}

func encodeRangeLimitsDescriptor(entry *DetailedTimingsRangeLimitsDescriptorEntry) ([detailedTimingDescriptorLength]byte, error) {
	if entry == nil {
		return [detailedTimingDescriptorLength]byte{}, fmt.Errorf("nil range limits descriptor")
	}

	if err := entry.Validate(); err != nil {
		return [detailedTimingDescriptorLength]byte{}, err
	}

	var descriptor [detailedTimingDescriptorLength]byte

	descriptor[3] = detailedTimingDescriptorTagRangeLimits
	descriptor[5] = entry.MinVerticalHz
	descriptor[6] = entry.MaxVerticalHz
	descriptor[7] = entry.MinKHz
	descriptor[8] = entry.MaxKHz
	descriptor[9] = entry.MaxClockMHz

	defaultGTFParameters := []byte{0x00, 0x0A, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20}
	copy(descriptor[10:], defaultGTFParameters)

	return descriptor, nil
}

func encodeMonitorTextDescriptor(tag byte, value string) ([detailedTimingDescriptorLength]byte, error) {
	descriptorPayload, err := encodeTextDescriptor(value)
	if err != nil {
		return [detailedTimingDescriptorLength]byte{}, err
	}

	var descriptor [detailedTimingDescriptorLength]byte

	descriptor[3] = tag
	copy(descriptor[5:], descriptorPayload[:])

	return descriptor, nil
}

func decode12Bit(low byte, highNibble byte) uint16 {
	return uint16(low) | (uint16(highNibble&0x0F) << 8)
}

func decode10Bit(low byte, highBits byte) uint16 {
	return uint16(low) | (uint16(highBits&0x03) << 8)
}

func decode6Bit(lowNibble byte, highBits byte) uint16 {
	return uint16(lowNibble&0x0F) | (uint16(highBits&0x03) << 4)
}

func split10Bit(value uint16) (byte, byte) {
	return byte(value & 0xFF), byte((value >> 8) & 0x03)
}

func split6Bit(value uint16) (byte, byte) {
	return byte(value & 0x0F), byte((value >> 4) & 0x03)
}

func isZeroDescriptor(descriptor []byte) bool {
	for _, value := range descriptor {
		if value != 0 {
			return false
		}
	}

	return true
}

func isBlankDescriptorPayload(payload []byte) bool {
	trimmed := bytes.Trim(payload, " \x00")
	return len(trimmed) == 0
}

func encodeTextDescriptor(value string) ([detailedTimingDescriptorTextDataLength]byte, error) {
	if strings.ContainsRune(value, '\n') {
		return [detailedTimingDescriptorTextDataLength]byte{}, fmt.Errorf("text descriptor must not contain a newline")
	}

	if len(value) > detailedTimingDescriptorTextMaxPayload {
		return [detailedTimingDescriptorTextDataLength]byte{}, fmt.Errorf("text descriptor too long: %d", len(value))
	}

	for _, runeValue := range value {
		if runeValue > 0x7F {
			return [detailedTimingDescriptorTextDataLength]byte{}, fmt.Errorf("text descriptor contains non-ASCII characters")
		}
	}

	var payload [detailedTimingDescriptorTextDataLength]byte

	copy(payload[:], value)
	payload[len(value)] = '\n'
	for index := len(value) + 1; index < len(payload); index++ {
		payload[index] = 0x20
	}

	return payload, nil
}

func decodeTextDescriptor(payload []byte) (string, error) {
	if len(payload) != detailedTimingDescriptorTextDataLength {
		return "", fmt.Errorf("invalid text descriptor length: %d", len(payload))
	}

	trimmed := bytes.TrimRight(payload, " \x00")

	if len(trimmed) == 0 {
		return "", fmt.Errorf("text descriptor empty")
	}

	if trimmed[len(trimmed)-1] == '\n' {
		trimmed = trimmed[:len(trimmed)-1]
	}

	for _, value := range trimmed {
		if value < 0x20 || value > 0x7E {
			return "", fmt.Errorf("text descriptor contains non-printable byte 0x%02x", value)
		}
	}

	return string(trimmed), nil
}
