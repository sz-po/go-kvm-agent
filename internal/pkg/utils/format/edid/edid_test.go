package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSpecificationFromBlock_RoundTrip(t *testing.T) {
	t.Parallel()

	testSpecification := Specification{
		Vendor: VendorSpecification{
			Manufacturer:      "ACM",
			ProductCode:       "1A2B",
			SerialNumber:      "00ABCDEF",
			WeekOfManufacture: 12,
			YearOfManufacture: 2021,
		},
		Display: DisplaySpecification{
			Input: DisplayInputSpecification{
				Digital: &DisplayDigitalInputSpecification{
					ColorBitDepth: uint8Ptr(8),
					Interface:     digitalInterfacePtr(DisplayPortDigitalInterface),
				},
			},
			Size: DisplaySizeSpecification{
				Width:  intPtr(52),
				Height: intPtr(32),
			},
			Gamma: float64Ptr(2.2),
			Features: DisplayFeaturesSpecification{
				SupportsStandby:            true,
				SupportsSuspend:            true,
				IsRgbColor:                 true,
				UsesStandardSrgbColorSpace: true,
				HasPreferredTimingMode:     true,
			},
		},
		Chromacity: ChromacitySpecification{
			RedX:   0.64,
			RedY:   0.33,
			GreenX: 0.30,
			GreenY: 0.60,
			BlueX:  0.15,
			BlueY:  0.06,
			WhiteX: 0.31,
			WhiteY: 0.33,
		},
		Timings: TimingsSpecification{
			Established: EstablishedTimingsSpecification{
				Supports640x480x60: true,
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
						MonitorName: &DetailedTimingsMonitorNameDescriptorEntry{Name: "RoundTrip"},
					},
				},
			},
		},
		ExtensionBlockCount: 1,
	}

	block, err := CreateBlockFromSpecification(testSpecification)
	assert.NoError(t, err)
	if assert.NotNil(t, block) {
		assert.Equal(t, byte(1), (*block)[edidExtensionIndex])
		assert.Equal(t, byte(0), calculateChecksum((*block)[:]))
	}

	decodedSpecification, decodeErr := CreateSpecificationFromBlock(*block)
	assert.NoError(t, decodeErr)
	if assert.NotNil(t, decodedSpecification) {
		assert.Equal(t, testSpecification.Vendor, decodedSpecification.Vendor)
		assert.Equal(t, testSpecification.ExtensionBlockCount, decodedSpecification.ExtensionBlockCount)
		assert.True(t, decodedSpecification.Display.Features.IsRgbColor)
		assert.NotNil(t, decodedSpecification.Display.Input.Digital)
		if decodedSpecification.Display.Input.Digital != nil {
			assert.NotNil(t, decodedSpecification.Display.Input.Digital.ColorBitDepth)
			if decodedSpecification.Display.Input.Digital.ColorBitDepth != nil {
				assert.Equal(t, uint8(8), *decodedSpecification.Display.Input.Digital.ColorBitDepth)
			}
			assert.NotNil(t, decodedSpecification.Display.Input.Digital.Interface)
			if decodedSpecification.Display.Input.Digital.Interface != nil {
				assert.Equal(t, string(DisplayPortDigitalInterface), string(*decodedSpecification.Display.Input.Digital.Interface))
			}
		}
		if assert.NotNil(t, decodedSpecification.Display.Gamma) {
			assert.InDelta(t, 2.2, *decodedSpecification.Display.Gamma, 0.0001)
		}
		assert.True(t, decodedSpecification.Timings.Established.Supports640x480x60)
		assert.Len(t, decodedSpecification.Timings.Standard.Entries, 1)
		assert.Equal(t, "RoundTrip", decodedSpecification.Timings.Detailed.Entries[0].MonitorName.Name)
		assert.InEpsilon(t, testSpecification.Chromacity.RedX, decodedSpecification.Chromacity.RedX, 0.001)
	}
}

func TestCreateSpecificationFromBlock_InvalidChecksum(t *testing.T) {
	t.Parallel()

	block, err := CreateBlockFromSpecification(validSpecification())
	assert.NoError(t, err)
	if !assert.NotNil(t, block) {
		return
	}

	(*block)[0]++

	_, decodeErr := CreateSpecificationFromBlock(*block)
	assert.Error(t, decodeErr)
	assert.Contains(t, decodeErr.Error(), "checksum")
}

func TestCreateSpecificationFromBlock_InvalidHeader(t *testing.T) {
	t.Parallel()

	block, err := CreateBlockFromSpecification(validSpecification())
	assert.NoError(t, err)
	if !assert.NotNil(t, block) {
		return
	}

	(*block)[edidHeaderStart] = 0xAA
	(*block)[edidChecksumIndex] = calculateChecksumByte((*block)[:edidChecksumIndex])

	_, decodeErr := CreateSpecificationFromBlock(*block)
	assert.Error(t, decodeErr)
	assert.Contains(t, decodeErr.Error(), "header")
}

func TestCreateSpecificationFromBlock_InvalidVersion(t *testing.T) {
	t.Parallel()

	block, err := CreateBlockFromSpecification(validSpecification())
	assert.NoError(t, err)
	if !assert.NotNil(t, block) {
		return
	}

	(*block)[edidVersionStart] = 0x02
	(*block)[edidChecksumIndex] = calculateChecksumByte((*block)[:edidChecksumIndex])

	_, decodeErr := CreateSpecificationFromBlock(*block)
	assert.Error(t, decodeErr)
	assert.Contains(t, decodeErr.Error(), "version")
}

func TestCreateBlockFromSpecification_InvalidVendor(t *testing.T) {
	t.Parallel()

	invalidSpecification := validSpecification()
	invalidSpecification.Vendor.Manufacturer = "Ac1"

	block, err := CreateBlockFromSpecification(invalidSpecification)
	assert.Nil(t, block)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vendor")
}

func TestSpecificationValidate(t *testing.T) {
	t.Parallel()

	var specification *Specification
	assert.EqualError(t, specification.Validate(), "nil specification")

	instance := validSpecification()
	instance.ExtensionBlockCount = 200
	assert.EqualError(t, instance.Validate(), "invalid extension block count")
}

func validSpecification() Specification {
	return Specification{
		Vendor: VendorSpecification{
			Manufacturer:      "ACM",
			ProductCode:       "1A2B",
			SerialNumber:      "00ABCDEF",
			WeekOfManufacture: 10,
			YearOfManufacture: 2020,
		},
		Display: DisplaySpecification{
			Input: DisplayInputSpecification{
				Digital: &DisplayDigitalInputSpecification{
					ColorBitDepth: uint8Ptr(8),
					Interface:     digitalInterfacePtr(DVIDigitalInterface),
				},
			},
			Size: DisplaySizeSpecification{
				Width:  intPtr(40),
				Height: intPtr(30),
			},
			Gamma: float64Ptr(2.2),
			Features: DisplayFeaturesSpecification{
				SupportsStandby:            true,
				IsRgbColor:                 true,
				UsesStandardSrgbColorSpace: true,
				HasPreferredTimingMode:     true,
			},
		},
		Chromacity: ChromacitySpecification{
			RedX:   0.64,
			RedY:   0.33,
			GreenX: 0.30,
			GreenY: 0.60,
			BlueX:  0.15,
			BlueY:  0.06,
			WhiteX: 0.31,
			WhiteY: 0.33,
		},
		Timings: TimingsSpecification{
			Established: EstablishedTimingsSpecification{
				Supports640x480x60: true,
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
						MonitorName: &DetailedTimingsMonitorNameDescriptorEntry{Name: "Valid"},
					},
				},
			},
		},
	}
}
