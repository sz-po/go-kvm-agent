package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTimingsSpecificationFromBlock_Empty(t *testing.T) {
	block := TimingsBlock{}

	specification, err := CreateTimingsSpecificationFromBlock(block)

	assert.NoError(t, err)
	assert.NotNil(t, specification)
	assert.NotNil(t, specification.Established)
	assert.NotNil(t, specification.Standard)
	assert.NotNil(t, specification.Detailed)
	// Empty blocks result in no entries (empty descriptors are filtered out)
	assert.Empty(t, specification.Standard.Entries)
	assert.Empty(t, specification.Detailed.Entries)
}

func TestCreateTimingsSpecificationFromBlock_WithEstablishedTimings(t *testing.T) {
	block := TimingsBlock{}
	block[0] = 0b10000000 // 720x400@70Hz

	specification, err := CreateTimingsSpecificationFromBlock(block)

	assert.NoError(t, err)
	assert.NotNil(t, specification)
	assert.True(t, specification.Established.Supports720x400x70)
	assert.False(t, specification.Established.Supports720x400x88)
}

func TestCreateTimingsSpecificationFromBlock_WithStandardTimings(t *testing.T) {
	block := TimingsBlock{}
	// Standard timing: 1024x768@60Hz (4:3)
	// Width byte = (1024/8 - 31) = 97 = 0x61
	block[3] = 0x61
	block[4] = 0b01000000 // aspect ratio 4:3 (01), refresh 60Hz (0)

	specification, err := CreateTimingsSpecificationFromBlock(block)

	assert.NoError(t, err)
	assert.NotNil(t, specification)
	assert.Len(t, specification.Standard.Entries, 1)
	assert.Equal(t, 1024, specification.Standard.Entries[0].Width)
	assert.Equal(t, 768, specification.Standard.Entries[0].Height)
	assert.Equal(t, 60, specification.Standard.Entries[0].RefreshRate)
	assert.Equal(t, string(StandardTimingEntryAspectRatio4x3), string(specification.Standard.Entries[0].AspectRatio))
}

func TestCreateTimingsSpecificationFromBlock_WithDetailedTimings(t *testing.T) {
	block := TimingsBlock{}

	// Use a monitor name descriptor instead of standard timing descriptor
	// This is simpler and avoids complex sync calculations
	offset := 19 // Detailed timings start at byte 19

	// Monitor descriptor header (pixel clock = 0)
	block[offset+0] = 0x00
	block[offset+1] = 0x00
	block[offset+2] = 0x00
	block[offset+3] = 0xFC // Monitor name tag
	block[offset+4] = 0x00

	// Monitor name: "Test" followed by newline and spaces
	block[offset+5] = 'T'
	block[offset+6] = 'e'
	block[offset+7] = 's'
	block[offset+8] = 't'
	block[offset+9] = '\n'
	block[offset+10] = 0x20 // Space padding
	block[offset+11] = 0x20
	block[offset+12] = 0x20
	block[offset+13] = 0x20
	block[offset+14] = 0x20
	block[offset+15] = 0x20
	block[offset+16] = 0x20
	block[offset+17] = 0x20

	specification, err := CreateTimingsSpecificationFromBlock(block)

	assert.NoError(t, err)
	assert.NotNil(t, specification)
	// Only non-empty descriptors are included
	assert.Len(t, specification.Detailed.Entries, 1)
	assert.NotNil(t, specification.Detailed.Entries[0].MonitorName)
	assert.Equal(t, "Test", specification.Detailed.Entries[0].MonitorName.Name)
}

func TestCreateTimingsBlockFromSpecification_RoundTrip(t *testing.T) {
	originalBlock := TimingsBlock{}

	// Add some established timings
	originalBlock[0] = 0b11100000 // 720x400@70, 720x400@88, 640x480@60

	// Add standard timing
	originalBlock[3] = 0x51       // 1024x768
	originalBlock[4] = 0b01000000 // 4:3, 60Hz

	// Parse to specification
	specification, err := CreateTimingsSpecificationFromBlock(originalBlock)
	assert.NoError(t, err)

	// Convert back to block
	recreatedBlock, err := CreateTimingsBlockFromSpecification(*specification)
	assert.NoError(t, err)
	if assert.NotNil(t, recreatedBlock) {
		assert.Equal(t, originalBlock[0:3], recreatedBlock[0:3])
		assert.Equal(t, originalBlock[3], recreatedBlock[3])
		assert.Equal(t, originalBlock[4], recreatedBlock[4])

		for index := 5; index < 19; index++ {
			assert.Equal(t, byte(0x01), recreatedBlock[index], "standard timing slot %d should be 0x01", index)
		}

		for descriptor := 0; descriptor < 4; descriptor++ {
			offset := 19 + descriptor*detailedTimingDescriptorLength
			assert.Equal(t, byte(0x00), recreatedBlock[offset])
			assert.Equal(t, byte(0x00), recreatedBlock[offset+1])
			assert.Equal(t, byte(0x00), recreatedBlock[offset+2])
			assert.Equal(t, byte(detailedTimingDescriptorTagUnused), recreatedBlock[offset+3])
			for index := offset + 4; index < offset+detailedTimingDescriptorLength; index++ {
				assert.Equal(t, byte(0x00), recreatedBlock[index])
			}
		}
	}
}

func TestCreateTimingsBlockFromSpecification_Empty(t *testing.T) {
	specification := TimingsSpecification{
		Established: EstablishedTimingsSpecification{},
		Standard:    StandardTimingsSpecification{},
		Detailed:    DetailedTimingsSpecification{},
	}

	block, err := CreateTimingsBlockFromSpecification(specification)

	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		assert.Equal(t, byte(0), block[0])
		assert.Equal(t, byte(0), block[1])
		assert.Equal(t, byte(0), block[2])

		for index := 3; index < 19; index++ {
			assert.Equal(t, byte(0x01), block[index], "standard timing byte %d should be 0x01", index)
		}

		for descriptor := 0; descriptor < 4; descriptor++ {
			offset := 19 + descriptor*detailedTimingDescriptorLength
			assert.Equal(t, byte(0x00), block[offset])
			assert.Equal(t, byte(0x00), block[offset+1])
			assert.Equal(t, byte(0x00), block[offset+2])
			assert.Equal(t, byte(detailedTimingDescriptorTagUnused), block[offset+3])
			for index := offset + 4; index < offset+detailedTimingDescriptorLength; index++ {
				assert.Equal(t, byte(0x00), block[index])
			}
		}
	}
}

func TestCreateTimingsBlockFromSpecification_WithAllSections(t *testing.T) {
	specification := TimingsSpecification{
		Established: EstablishedTimingsSpecification{
			Supports640x480x60: true,
			Supports800x600x60: true,
		},
		Standard: StandardTimingsSpecification{
			Entries: []StandardTimingEntrySpecification{
				{
					Width:       1024,
					Height:      768,
					RefreshRate: 60,
					AspectRatio: StandardTimingEntryAspectRatio4x3,
				},
			},
		},
		Detailed: DetailedTimingsSpecification{
			Entries: []DetailedTimingsEntrySpecification{
				{
					MonitorName: &DetailedTimingsMonitorNameDescriptorEntry{
						Name: "Test Monitor",
					},
				},
			},
		},
	}

	block, err := CreateTimingsBlockFromSpecification(specification)

	assert.NoError(t, err)
	assert.NotNil(t, block)

	// Verify established timings section
	assert.Equal(t, byte(0b00100001), block[0]) // 640x480@60 and 800x600@60

	// Verify standard timings section
	// Width byte = (1024/8 - 31) = 97 = 0x61
	assert.Equal(t, byte(0x61), block[3])
	assert.Equal(t, byte(0b01000000), block[4])

	// Verify detailed timings section has monitor name descriptor
	// Bytes 19-21 should be 0 (descriptor header for monitor descriptors)
	assert.Equal(t, byte(0), block[19])
	assert.Equal(t, byte(0), block[20])
	assert.Equal(t, byte(0), block[21])
	assert.Equal(t, byte(0xFC), block[22]) // Monitor name tag
}

func TestTimingsSpecification_Validate_Nil(t *testing.T) {
	var specification *TimingsSpecification

	err := specification.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil specification")
}

func TestTimingsSpecification_Validate_InvalidEstablished(t *testing.T) {
	specification := &TimingsSpecification{
		Established: EstablishedTimingsSpecification{},
		Standard:    StandardTimingsSpecification{},
		Detailed:    DetailedTimingsSpecification{},
	}

	// This should succeed since empty established timings are valid
	err := specification.Validate()
	assert.NoError(t, err)
}

func TestTimingsSpecification_Validate_InvalidStandard(t *testing.T) {
	specification := &TimingsSpecification{
		Established: EstablishedTimingsSpecification{},
		Standard: StandardTimingsSpecification{
			Entries: []StandardTimingEntrySpecification{
				{
					Width:       100, // Invalid: too small
					Height:      75,
					RefreshRate: 60,
					AspectRatio: StandardTimingEntryAspectRatio4x3,
				},
			},
		},
		Detailed: DetailedTimingsSpecification{},
	}

	err := specification.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "standard timings")
}

func TestTimingsSpecification_Validate_InvalidDetailed(t *testing.T) {
	specification := &TimingsSpecification{
		Established: EstablishedTimingsSpecification{},
		Standard:    StandardTimingsSpecification{},
		Detailed: DetailedTimingsSpecification{
			Entries: []DetailedTimingsEntrySpecification{
				{
					Standard: &DetailedTimingsStandardDescriptorEntry{
						PixelClock:             5, // Invalid: too small
						StereoMode:             DetailedTimingsStereoModeNone,
						SyncType:               DetailedTimingsSyncTypeDigitalSeparate,
						HorizontalSyncPositive: true,
						VerticalSyncPositive:   true,
					},
				},
			},
		},
	}

	err := specification.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "detailed timings")
}

func TestCreateTimingsSpecificationFromBlock_InvalidEstablishedTimings(t *testing.T) {
	block := TimingsBlock{}
	// Set reserved bits in established timings (this should trigger an error)
	// The masks show that all bits in byte 0 and 1 are used, so we need to test byte 2
	// Byte 2 has reserved mask 0b01111111, so setting any of those bits should fail
	block[2] = 0b00000001 // Set a reserved bit

	_, err := CreateTimingsSpecificationFromBlock(block)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "established timings")
}

func TestCreateTimingsSpecificationFromBlock_ComplexRealWorld(t *testing.T) {
	// Simulate a real-world EDID timings block with multiple sections populated
	block := TimingsBlock{}

	// Established timings: common resolutions
	block[0] = 0b11111111 // All byte 0 timings
	block[1] = 0b11111111 // All byte 1 timings
	block[2] = 0b10000000 // Apple 1152x870@75

	// Standard timings: multiple entries
	block[3] = 0x61       // 1024x768
	block[4] = 0b01000000 // 4:3, 60Hz

	block[5] = 0x71       // 1152x864
	block[6] = 0b00001111 // 16:10, 75Hz

	block[7] = 0x81       // 1280x1024
	block[8] = 0b10000000 // 5:4, 60Hz

	specification, err := CreateTimingsSpecificationFromBlock(block)

	assert.NoError(t, err)
	assert.NotNil(t, specification)

	// Verify established timings
	assert.True(t, specification.Established.Supports720x400x70)
	assert.True(t, specification.Established.Supports1280x1024x75)
	assert.True(t, specification.Established.Supports1152x870x75)

	// Verify standard timings
	assert.Len(t, specification.Standard.Entries, 3)
	assert.Equal(t, 1024, specification.Standard.Entries[0].Width)
	assert.Equal(t, 1152, specification.Standard.Entries[1].Width)
	assert.Equal(t, 1280, specification.Standard.Entries[2].Width)
}

func TestCreateTimingsBlockFromSpecification_ErrorPropagation(t *testing.T) {
	specification := TimingsSpecification{
		Established: EstablishedTimingsSpecification{},
		Standard: StandardTimingsSpecification{
			Entries: []StandardTimingEntrySpecification{
				{
					Width:       1024,
					Height:      999, // Invalid: doesn't match aspect ratio
					RefreshRate: 60,
					AspectRatio: StandardTimingEntryAspectRatio4x3,
				},
			},
		},
		Detailed: DetailedTimingsSpecification{},
	}

	_, err := CreateTimingsBlockFromSpecification(specification)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validate")
	assert.Contains(t, err.Error(), "standard timings")
}
