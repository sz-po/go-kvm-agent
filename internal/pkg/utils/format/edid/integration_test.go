package edid

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSpecificationFromBlockIntegration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		hexEDID      string
		expectedSpec Specification
	}{}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rawBytes := decodeIntegrationHex(t, testCase.hexEDID)
			require.Len(t, rawBytes, edidBlockLength, "EDID block must be exactly 128 bytes.")
			assert.Equal(t, byte(0), calculateChecksum(rawBytes), "EDID checksum must equal zero.")

			var block Block
			copy(block[:], rawBytes)

			specification, err := CreateSpecificationFromBlock(block)
			require.NoError(t, err)
			require.NotNil(t, specification)

			defer func() {
				if specification != nil {
					if actualBlock, err := CreateBlockFromSpecification(*specification); err == nil {
						saveEDIDOnFailure(t, testCase.name, rawBytes, actualBlock[:])
					}
				}
			}()

			assert.Equal(t, testCase.expectedSpec, *specification)
		})
	}
}

func TestCreateBlockFromSpecificationIntegration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		specification   Specification
		expectedHexEDID string
	}{
		{
			name: "1080p@30",
			specification: Specification{
				Vendor: VendorSpecification{
					Manufacturer:      "ACM",
					ProductCode:       "1A2B",
					SerialNumber:      "12345678",
					WeekOfManufacture: 12,
					YearOfManufacture: 2021,
				},
				Display: DisplaySpecification{
					Input: DisplayInputSpecification{
						Digital: &DisplayDigitalInputSpecification{
							Interface: digitalInterfacePtr(DVIDigitalInterface),
						},
					},
					Size: DisplaySizeSpecification{
						Width:  intPtr(32),
						Height: intPtr(18),
					},
					Features: DisplayFeaturesSpecification{
						SupportsStandby:            true,
						SupportsSuspend:            true,
						SupportsActiveOff:          true,
						IsRgbColor:                 true,
						UsesStandardSrgbColorSpace: true,
						HasPreferredTimingMode:     true,
					},
					Gamma: float64Ptr(2.2),
				},
				Chromacity: ChromacitySpecification{
					RedX:   0.6400,
					RedY:   0.3300,
					GreenX: 0.3000,
					GreenY: 0.6000,
					BlueX:  0.1500,
					BlueY:  0.0600,
					WhiteX: 0.3127,
					WhiteY: 0.3290,
				},
				Timings: TimingsSpecification{
					Detailed: DetailedTimingsSpecification{
						Entries: []DetailedTimingsEntrySpecification{
							{
								Standard: &DetailedTimingsStandardDescriptorEntry{
									HorizontalActive:     1920,
									HorizontalBlank:      280,
									HorizontalSyncOffset: 88,
									HorizontalSyncWidth:  44,
									HorizontalImageSize:  320,
									HorizontalBorder:     0,
									VerticalActive:       1080,
									VerticalBlank:        45,
									VerticalSyncOffset:   4,
									VerticalSyncWidth:    5,
									VerticalImageSize:    180,
									VerticalBorder:       0,
									PixelClock:           74250,

									HorizontalSyncPositive: true,
									VerticalSyncPositive:   true,
									SyncType:               DetailedTimingsSyncTypeDigitalSeparate,
									StereoMode:             DetailedTimingsStereoModeNone,
								},
							},
							{
								RangeLimits: &DetailedTimingsRangeLimitsDescriptorEntry{
									MinVerticalHz: 24,
									MaxVerticalHz: 75,
									MinKHz:        30,
									MaxKHz:        83,
									MaxClockMHz:   15,
								},
							},
							{
								MonitorSerial: &DetailedTimingsMonitorSerialEntry{
									Name: "XXXXYYYY",
								},
							},
							{
								MonitorName: &DetailedTimingsMonitorNameDescriptorEntry{
									Name: "Test",
								},
							},
						},
					},
				},
			},
			expectedHexEDID: "00ffffffffffff00046d1a2b123456780c1f0103812012786fee91a3544c99260f505400000001010101010101010101010101010101011d801871382d40582c450040b41000001e000000fd00184b1e530f000a202020202020000000ff0058585858595959590a20202020000000fc00546573740a202020202020202000db",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			block, err := CreateBlockFromSpecification(testCase.specification)
			require.NoError(t, err)
			require.NotNil(t, block)
			assert.Equal(t, byte(0), calculateChecksum(block[:]), "Generated EDID checksum must equal zero.")

			expectedBytes := decodeIntegrationHex(t, testCase.expectedHexEDID)
			require.Len(t, expectedBytes, edidBlockLength, "Expected EDID block must be exactly 128 bytes.")

			defer saveEDIDOnFailure(t, testCase.name, expectedBytes, block[:])

			assert.Equal(t, expectedBytes, block[:])
		})
	}
}

func decodeIntegrationHex(t *testing.T, value string) []byte {
	t.Helper()

	sanitized := sanitizeHexString(value)
	require.Condition(t, func() bool { return len(sanitized)%2 == 0 }, "hex string must contain an even number of digits")

	decoded, err := hex.DecodeString(sanitized)
	require.NoError(t, err, "decode EDID hex string")

	return decoded
}

func sanitizeHexString(value string) string {
	builder := make([]rune, 0, len(value))

	for _, char := range value {
		if (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F') {
			builder = append(builder, char)
		}
	}

	return string(builder)
}

func saveEDIDOnFailure(t *testing.T, testName string, expected, actual []byte) {
	t.Helper()

	if !t.Failed() {
		return
	}

	expectedPath := testName + "-expected.edid"
	if err := os.WriteFile(expectedPath, expected, 0644); err != nil {
		t.Logf("Failed to write expected EDID to %s: %v", expectedPath, err)
	} else {
		t.Logf("Saved expected EDID to %s", expectedPath)
	}

	actualPath := testName + "-actual.edid"
	if err := os.WriteFile(actualPath, actual, 0644); err != nil {
		t.Logf("Failed to write actual EDID to %s: %v", actualPath, err)
	} else {
		t.Logf("Saved actual EDID to %s", actualPath)
	}
}
