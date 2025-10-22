package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func validStandardDescriptorEntry() DetailedTimingsStandardDescriptorEntry {
	return DetailedTimingsStandardDescriptorEntry{
		PixelClock:             74250,
		HorizontalActive:       1280,
		HorizontalBlank:        370,
		VerticalActive:         720,
		VerticalBlank:          30,
		HorizontalSyncOffset:   48,
		HorizontalSyncWidth:    32,
		VerticalSyncOffset:     3,
		VerticalSyncWidth:      5,
		HorizontalImageSize:    510,
		VerticalImageSize:      287,
		HorizontalBorder:       0,
		VerticalBorder:         0,
		Interlaced:             false,
		StereoMode:             DetailedTimingsStereoModeNone,
		SyncType:               DetailedTimingsSyncTypeDigitalSeparate,
		HorizontalSyncPositive: true,
		VerticalSyncPositive:   true,
	}
}

func TestDetailedTimingsRoundTrip(t *testing.T) {
	specification := DetailedTimingsSpecification{
		Entries: []DetailedTimingsEntrySpecification{
			{
				Standard: func() *DetailedTimingsStandardDescriptorEntry {
					entry := validStandardDescriptorEntry()
					return &entry
				}(),
			},
			{
				RangeLimits: &DetailedTimingsRangeLimitsDescriptorEntry{
					MinVerticalHz: 50,
					MaxVerticalHz: 75,
					MinKHz:        30,
					MaxKHz:        80,
					MaxClockMHz:   150,
				},
			},
			{
				MonitorName: &DetailedTimingsMonitorNameDescriptorEntry{Name: "MyDisplay"},
			},
		},
	}

	block, err := CreateDetailedTimingsBlockFromSpecification(specification)
	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		roundTripSpecification, roundTripErr := CreateDetailedTimingsSpecificationFromBlock(*block)
		assert.NoError(t, roundTripErr)
		if assert.NotNil(t, roundTripSpecification) {
			assert.Equal(t, specification.Entries, roundTripSpecification.Entries)
		}
	}
}

func TestCreateDetailedTimingsSpecificationFromBlockUnsupportedSyncBits(t *testing.T) {
	specification := DetailedTimingsSpecification{
		Entries: []DetailedTimingsEntrySpecification{
			{
				Standard: func() *DetailedTimingsStandardDescriptorEntry {
					entry := validStandardDescriptorEntry()
					return &entry
				}(),
			},
		},
	}

	block, err := CreateDetailedTimingsBlockFromSpecification(specification)
	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		(*block)[17] |= 0x01

		roundTripSpecification, roundTripErr := CreateDetailedTimingsSpecificationFromBlock(*block)
		assert.Nil(t, roundTripSpecification)
		assert.EqualError(t, roundTripErr, "decode descriptor 0: unsupported digital separate sync reserved bit: 0x01")
	}
}

func TestCreateDetailedTimingsSpecificationFromBlockUnsupportedMonitorDescriptorTag(t *testing.T) {
	var block DetailedTimingsBlock

	block[3] = detailedTimingDescriptorTagAsciiString
	block[5] = 'T'

	specification, err := CreateDetailedTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "decode descriptor 0: ascii string descriptors are not supported")
}

func TestCreateDetailedTimingsSpecificationFromBlockRangeExtensionError(t *testing.T) {
	var block DetailedTimingsBlock

	block[3] = detailedTimingDescriptorTagRangeLimits
	block[5] = 40
	block[6] = 70
	block[7] = 30
	block[8] = 80
	block[9] = 150
	block[10] = 1

	specification, err := CreateDetailedTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "decode descriptor 0: unsupported range descriptor extension")
}

func TestCreateDetailedTimingsSpecificationFromBlockUnexpectedMonitorDescriptorHeader(t *testing.T) {
	var block DetailedTimingsBlock

	block[2] = 1
	block[3] = detailedTimingDescriptorTagMonitorName

	specification, err := CreateDetailedTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "decode descriptor 0: unexpected monitor descriptor header")
}

func TestCreateDetailedTimingsSpecificationFromBlockBlankUnusedDescriptor(t *testing.T) {
	var block DetailedTimingsBlock

	block[3] = detailedTimingDescriptorTagUnused
	for index := 5; index < 5+detailedTimingDescriptorTextDataLength; index++ {
		block[index] = 0x20
	}

	specification, err := CreateDetailedTimingsSpecificationFromBlock(block)

	assert.NoError(t, err)
	if assert.NotNil(t, specification) {
		assert.Empty(t, specification.Entries)
	}
}

func TestCreateDetailedTimingsSpecificationFromBlockInvalidStandardDescriptor(t *testing.T) {
	var block DetailedTimingsBlock

	block[0] = 0x01 // pixel clock 74250 kHz -> 7425 * 10kHz units
	block[1] = 0x1D
	block[2] = 0x00 // horizontal active low (1280)
	block[3] = 0x46 // horizontal blank low (70)
	block[4] = 0x50 // high nibbles: active=0x5, blank=0x0
	block[5] = 0xD0 // vertical active low (720)
	block[6] = 0x1E // vertical blank low (30)
	block[7] = 0x20 // high nibbles: active=0x2, blank=0x0
	block[8] = 0x3C // horizontal sync offset low (60)
	block[9] = 0x20 // horizontal sync width low (32)
	block[10] = 0x00
	block[11] = 0x35 // vertical sync offset/width (3/5)
	block[12] = 0xFE // horizontal image low (510)
	block[13] = 0x1F // vertical image low (287)
	block[14] = 0x11 // high nibbles for image sizes
	// borders and flags are already zero

	specification, err := CreateDetailedTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "decode descriptor 0: horizontal sync exceeds blanking interval")
}

func TestCreateDetailedTimingsBlockFromSpecificationTooManyEntries(t *testing.T) {
	specification := DetailedTimingsSpecification{}

	for index := 0; index < detailedTimingDescriptorCount+1; index++ {
		specification.Entries = append(specification.Entries, DetailedTimingsEntrySpecification{})
	}

	block, err := CreateDetailedTimingsBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "validate specification: too many detailed timing entries: 5")
}

func TestDetailedTimingsEntrySpecificationValidateMultipleDescriptors(t *testing.T) {
	entry := &DetailedTimingsEntrySpecification{
		Standard: func() *DetailedTimingsStandardDescriptorEntry {
			entry := validStandardDescriptorEntry()
			return &entry
		}(),
		RangeLimits: &DetailedTimingsRangeLimitsDescriptorEntry{
			MinVerticalHz: 50,
			MaxVerticalHz: 75,
			MinKHz:        30,
			MaxKHz:        80,
			MaxClockMHz:   150,
		},
	}

	err := entry.Validate()

	assert.EqualError(t, err, "multiple descriptor types provided")
}

func TestCreateDetailedTimingsBlockFromSpecificationMonitorNameValidation(t *testing.T) {
	specification := DetailedTimingsSpecification{
		Entries: []DetailedTimingsEntrySpecification{
			{
				MonitorName: &DetailedTimingsMonitorNameDescriptorEntry{Name: "Invalid\nName"},
			},
		},
	}

	block, err := CreateDetailedTimingsBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "validate specification: entry 0 invalid: monitor name descriptor invalid: monitor name must not contain a newline")
}

func TestCreateDetailedTimingsSpecificationFromBlockEmptyTextDescriptor(t *testing.T) {
	var block DetailedTimingsBlock

	block[3] = detailedTimingDescriptorTagMonitorName

	for index := 5; index < 5+detailedTimingDescriptorTextDataLength; index++ {
		block[index] = 0x20
	}

	specification, err := CreateDetailedTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "decode descriptor 0: decode monitor name: text descriptor empty")
}

func TestCreateDetailedTimingsSpecificationFromBlockUnsupportedDescriptorLength(t *testing.T) {
	descriptor := make([]byte, detailedTimingDescriptorLength-1)

	_, err := decodeDetailedTimingDescriptor(descriptor)

	assert.EqualError(t, err, "descriptor length mismatch: 17")
}

func TestCreateDetailedTimingsSpecificationFromBlockMonitorSerial(t *testing.T) {
	specification := DetailedTimingsSpecification{
		Entries: []DetailedTimingsEntrySpecification{
			{
				MonitorSerial: &DetailedTimingsMonitorSerialEntry{Name: "SN123"},
			},
		},
	}

	block, err := CreateDetailedTimingsBlockFromSpecification(specification)
	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		roundTripSpecification, roundTripErr := CreateDetailedTimingsSpecificationFromBlock(*block)
		assert.NoError(t, roundTripErr)
		if assert.NotNil(t, roundTripSpecification) {
			assert.Equal(t, specification.Entries, roundTripSpecification.Entries)
		}
	}
}

func TestMonitorSerialValidationError(t *testing.T) {
	specification := DetailedTimingsSpecification{
		Entries: []DetailedTimingsEntrySpecification{
			{
				MonitorSerial: &DetailedTimingsMonitorSerialEntry{Name: "Bad\nSerial"},
			},
		},
	}

	block, err := CreateDetailedTimingsBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "validate specification: entry 0 invalid: monitor serial descriptor invalid: monitor serial must not contain a newline")
}

func TestMonitorNameTooLong(t *testing.T) {
	specification := DetailedTimingsSpecification{
		Entries: []DetailedTimingsEntrySpecification{
			{
				MonitorName: &DetailedTimingsMonitorNameDescriptorEntry{Name: "ABCDEFGHIJKLM"},
			},
		},
	}

	block, err := CreateDetailedTimingsBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "validate specification: entry 0 invalid: monitor name descriptor invalid: monitor name too long: 13")
}

func TestDetailedTimingsSpecificationValidateNil(t *testing.T) {
	var specification *DetailedTimingsSpecification

	err := specification.Validate()

	assert.EqualError(t, err, "nil specification")
}

func TestDecodeTextDescriptorNonPrintable(t *testing.T) {
	var block DetailedTimingsBlock

	block[3] = detailedTimingDescriptorTagMonitorName
	block[5] = 0x01
	block[6] = 0x0A
	block[7] = 0x20

	specification, err := CreateDetailedTimingsSpecificationFromBlock(block)

	assert.Nil(t, specification)
	assert.EqualError(t, err, "decode descriptor 0: decode monitor name: text descriptor contains non-printable byte 0x01")
}

func TestDetailedTimingsStandardDescriptorEntryValidatePixelClock(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.PixelClock = 74255

	err := entry.Validate()

	assert.EqualError(t, err, "pixel clock must be divisible by 10")
}

func TestDetailedTimingsStandardDescriptorEntryValidateVerticalSync(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.VerticalBlank = 10
	entry.VerticalSyncOffset = 6
	entry.VerticalSyncWidth = 6

	err := entry.Validate()

	assert.EqualError(t, err, "vertical sync exceeds blanking interval")
}

func TestDetailedTimingsStandardDescriptorEntryValidateStereoModeUnknown(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.StereoMode = DetailedTimingsStereoModeUnknown

	err := entry.Validate()

	assert.EqualError(t, err, "unsupported stereo mode: 0")
}

func TestDetailedTimingsStandardDescriptorEntryValidateSyncTypeUnknown(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.SyncType = DetailedTimingsSyncTypeUnknown

	err := entry.Validate()

	assert.EqualError(t, err, "unsupported sync type: 0")
}

func TestDetailedTimingsStandardDescriptorEntryValidateAnalogCompositeLocationMissing(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.SyncType = DetailedTimingsSyncTypeAnalogComposite
	entry.HorizontalSyncPositive = false
	entry.VerticalSyncPositive = false

	err := entry.Validate()

	assert.EqualError(t, err, "analog composite sync must set exactly one location flag")
}

func TestDetailedTimingsStandardDescriptorEntryValidateAnalogCompositeDigitalPolarity(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.SyncType = DetailedTimingsSyncTypeAnalogComposite
	entry.CompositeSyncOnGreenOnly = true

	err := entry.Validate()

	assert.EqualError(t, err, "analog composite sync cannot set digital polarity flags")
}

func TestDetailedTimingsStandardDescriptorEntryValidateAnalogBipolarLocationFlag(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.SyncType = DetailedTimingsSyncTypeAnalogBipolar
	entry.CompositeSyncOnGreenOnly = true
	entry.HorizontalSyncPositive = false
	entry.VerticalSyncPositive = false

	err := entry.Validate()

	assert.EqualError(t, err, "analog bipolar sync does not support composite sync location flags")
}

func TestDetailedTimingsStandardDescriptorEntryValidateAnalogBipolarSerration(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.SyncType = DetailedTimingsSyncTypeAnalogBipolar
	entry.CompositeSyncSerration = true
	entry.HorizontalSyncPositive = false
	entry.VerticalSyncPositive = false

	err := entry.Validate()

	assert.EqualError(t, err, "analog bipolar sync does not support composite sync serration")
}

func TestDetailedTimingsStandardDescriptorEntryValidateDigitalCompositeLocationFlag(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.SyncType = DetailedTimingsSyncTypeDigitalComposite
	entry.CompositeSyncOnAllColors = true

	err := entry.Validate()

	assert.EqualError(t, err, "digital composite sync does not support analog location flags")
}

func TestDetailedTimingsStandardDescriptorEntryValidateDigitalCompositePolarity(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.SyncType = DetailedTimingsSyncTypeDigitalComposite
	entry.HorizontalSyncPositive = true

	err := entry.Validate()

	assert.EqualError(t, err, "digital composite sync cannot set separate polarity flags")
}

func TestDetailedTimingsStandardDescriptorEntryValidateDigitalSeparateSerration(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.CompositeSyncSerration = true

	err := entry.Validate()

	assert.EqualError(t, err, "digital separate sync does not support composite serration")
}

func TestDetailedTimingsStandardDescriptorEntryValidateDigitalSeparateAnalogLocation(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.CompositeSyncOnGreenOnly = true

	err := entry.Validate()

	assert.EqualError(t, err, "digital separate sync does not support analog location flags")
}

func TestCreateDetailedTimingsBlockFromSpecificationHorizontalSyncValidation(t *testing.T) {
	specification := DetailedTimingsSpecification{
		Entries: []DetailedTimingsEntrySpecification{
			{
				Standard: func() *DetailedTimingsStandardDescriptorEntry {
					entry := validStandardDescriptorEntry()
					entry.HorizontalBlank = 70
					entry.HorizontalSyncOffset = 64
					entry.HorizontalSyncWidth = 20
					return &entry
				}(),
			},
		},
	}

	block, err := CreateDetailedTimingsBlockFromSpecification(specification)

	assert.Nil(t, block)
	assert.EqualError(t, err, "validate specification: entry 0 invalid: standard descriptor invalid: horizontal sync exceeds blanking interval")
}

func TestDecodeDetailedTimingFlagsAnalogCompositeLocations(t *testing.T) {
	testCases := []struct {
		name              string
		flagValue         byte
		expectedSerration bool
		expectedGreenOnly bool
		expectedAllColors bool
		expectedRedBlue   bool
	}{
		{
			name:              "GreenOnly",
			flagValue:         0x00,
			expectedGreenOnly: true,
		},
		{
			name:              "AllColors",
			flagValue:         0x01,
			expectedAllColors: true,
		},
		{
			name:            "RedBlue",
			flagValue:       0x02,
			expectedRedBlue: true,
		},
		{
			name:              "WithSerration",
			flagValue:         0x04,
			expectedSerration: true,
			expectedGreenOnly: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			entry := validStandardDescriptorEntry()
			entry.HorizontalSyncPositive = false
			entry.VerticalSyncPositive = false

			err := decodeDetailedTimingFlags(testCase.flagValue, &entry)
			assert.NoError(t, err)
			assert.Equal(t, DetailedTimingsStereoModeNone, entry.StereoMode)
			assert.Equal(t, DetailedTimingsSyncTypeAnalogComposite, entry.SyncType)
			assert.Equal(t, testCase.expectedSerration, entry.CompositeSyncSerration)
			assert.Equal(t, testCase.expectedGreenOnly, entry.CompositeSyncOnGreenOnly)
			assert.Equal(t, testCase.expectedAllColors, entry.CompositeSyncOnAllColors)
			assert.Equal(t, testCase.expectedRedBlue, entry.CompositeSyncOnRedBlue)
			assert.False(t, entry.HorizontalSyncPositive)
			assert.False(t, entry.VerticalSyncPositive)
		})
	}
}

func TestDecodeDetailedTimingFlagsAnalogCompositeInvalidLocation(t *testing.T) {
	entry := validStandardDescriptorEntry()
	entry.HorizontalSyncPositive = false
	entry.VerticalSyncPositive = false

	err := decodeDetailedTimingFlags(0x03, &entry)

	assert.EqualError(t, err, "unsupported analog composite sync location bits: 0x03")
}

func TestDecodeDetailedTimingFlagsAnalogBipolarInvalidBits(t *testing.T) {
	entry := validStandardDescriptorEntry()

	err := decodeDetailedTimingFlags(0x09, &entry)

	assert.EqualError(t, err, "unsupported analog bipolar sync detail bits: 0x01")
}

func TestDecodeDetailedTimingFlagsDigitalCompositeInvalidBits(t *testing.T) {
	entry := validStandardDescriptorEntry()

	err := decodeDetailedTimingFlags(0x13, &entry)

	assert.EqualError(t, err, "unsupported digital composite sync detail bits: 0x03")
}

func TestEncodeDetailedTimingFlags(t *testing.T) {
	testCases := []struct {
		name      string
		configure func(entry *DetailedTimingsStandardDescriptorEntry)
		expected  byte
	}{
		{
			name: "DigitalSeparatePositivePolarity",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeDigitalSeparate
				entry.HorizontalSyncPositive = true
				entry.VerticalSyncPositive = true
			},
			expected: 0x1E,
		},
		{
			name: "DigitalSeparateNegativePolarity",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeDigitalSeparate
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
			},
			expected: 0x18,
		},
		{
			name: "AnalogCompositeGreenOnly",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeAnalogComposite
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.CompositeSyncOnGreenOnly = true
			},
			expected: 0x00,
		},
		{
			name: "AnalogCompositeAllColorsWithSerration",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeAnalogComposite
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.CompositeSyncOnAllColors = true
				entry.CompositeSyncSerration = true
			},
			expected: 0x05,
		},
		{
			name: "AnalogCompositeRedBlue",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeAnalogComposite
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.CompositeSyncOnRedBlue = true
			},
			expected: 0x02,
		},
		{
			name: "AnalogCompositeInterlaced",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeAnalogComposite
				entry.Interlaced = true
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.CompositeSyncOnGreenOnly = true
			},
			expected: 0x80,
		},
		{
			name: "AnalogBipolar",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeAnalogBipolar
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
			},
			expected: 0x08,
		},
		{
			name: "DigitalCompositeWithSerration",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeDigitalComposite
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.CompositeSyncSerration = true
			},
			expected: 0x14,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			entry := validStandardDescriptorEntry()
			testCase.configure(&entry)

			err := entry.Validate()
			assert.NoError(t, err)

			flagValue, encodeErr := encodeDetailedTimingFlags(&entry)
			assert.NoError(t, encodeErr)
			assert.Equal(t, testCase.expected, flagValue)
		})
	}
}

func TestDetailedTimingsRoundTripSyncVariants(t *testing.T) {
	testCases := []struct {
		name      string
		configure func(entry *DetailedTimingsStandardDescriptorEntry)
	}{
		{
			name: "AnalogCompositeGreenFieldSequentialRight",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeAnalogComposite
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.CompositeSyncOnGreenOnly = true
				entry.StereoMode = DetailedTimingsStereoModeFieldSequentialRight
			},
		},
		{
			name: "AnalogCompositeAllColorsFieldSequentialLeftSerration",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeAnalogComposite
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.CompositeSyncOnAllColors = true
				entry.CompositeSyncSerration = true
				entry.StereoMode = DetailedTimingsStereoModeFieldSequentialLeft
			},
		},
		{
			name: "AnalogCompositeRedBlueTwoWayInterleavedInterlaced",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeAnalogComposite
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.CompositeSyncOnRedBlue = true
				entry.Interlaced = true
				entry.StereoMode = DetailedTimingsStereoModeTwoWayInterleaved
			},
		},
		{
			name: "AnalogBipolarStereoNone",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeAnalogBipolar
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.StereoMode = DetailedTimingsStereoModeNone
			},
		},
		{
			name: "DigitalCompositeSerration",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeDigitalComposite
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.CompositeSyncSerration = true
				entry.StereoMode = DetailedTimingsStereoModeFieldSequentialLeft
			},
		},
		{
			name: "DigitalSeparateNegativePolarity",
			configure: func(entry *DetailedTimingsStandardDescriptorEntry) {
				entry.SyncType = DetailedTimingsSyncTypeDigitalSeparate
				entry.HorizontalSyncPositive = false
				entry.VerticalSyncPositive = false
				entry.StereoMode = DetailedTimingsStereoModeFieldSequentialLeft
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			entry := validStandardDescriptorEntry()
			testCase.configure(&entry)

			specification := DetailedTimingsSpecification{
				Entries: []DetailedTimingsEntrySpecification{
					{Standard: func() *DetailedTimingsStandardDescriptorEntry {
						entryCopy := entry
						return &entryCopy
					}()},
				},
			}

			block, err := CreateDetailedTimingsBlockFromSpecification(specification)
			assert.NoError(t, err)
			if assert.NotNil(t, block) {
				roundTripSpecification, decodeErr := CreateDetailedTimingsSpecificationFromBlock(*block)
				assert.NoError(t, decodeErr)
				if assert.NotNil(t, roundTripSpecification) {
					if assert.Len(t, roundTripSpecification.Entries, 1) {
						decoded := roundTripSpecification.Entries[0].Standard
						if assert.NotNil(t, decoded) {
							assert.Equal(t, entry, *decoded)
						}
					}
				}
			}
		})
	}
}
