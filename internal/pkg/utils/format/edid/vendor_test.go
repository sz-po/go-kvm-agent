package edid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVendorSpecificationValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		specification VendorSpecification
		wantErr       bool
		wantErrIs     error
		notErrIs      error
	}{
		{
			name: "ValidWeekWithinRange",
			specification: VendorSpecification{
				Manufacturer:      "HPQ",
				ProductCode:       "1A2B",
				SerialNumber:      "FE01DCBA",
				WeekOfManufacture: 10,
				YearOfManufacture: 2020,
			},
		},
		{
			name: "ValidWeekZero",
			specification: VendorSpecification{
				Manufacturer:      "HPQ",
				ProductCode:       "1A2B",
				SerialNumber:      "FE01DCBA",
				WeekOfManufacture: 0,
				YearOfManufacture: 2020,
			},
		},
		{
			name: "ValidWeekSentinel",
			specification: VendorSpecification{
				Manufacturer:      "HPQ",
				ProductCode:       "1A2B",
				SerialNumber:      "FE01DCBA",
				WeekOfManufacture: 255,
				YearOfManufacture: 2020,
			},
		},
		{
			name: "InvalidWeek",
			specification: VendorSpecification{
				Manufacturer:      "HPQ",
				ProductCode:       "1A2B",
				SerialNumber:      "FE01DCBA",
				WeekOfManufacture: 54,
				YearOfManufacture: 2020,
			},
			wantErr:   true,
			wantErrIs: ErrInvalidWeekOfManufacture,
		},
		{
			name: "InvalidManufacturer",
			specification: VendorSpecification{
				Manufacturer:      "hpq",
				ProductCode:       "1A2B",
				SerialNumber:      "FE01DCBA",
				WeekOfManufacture: 10,
				YearOfManufacture: 2020,
			},
			wantErr:  true,
			notErrIs: ErrInvalidWeekOfManufacture,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := testCase.specification.Validate()

			if testCase.wantErr {
				assert.Error(t, err)
				if testCase.wantErrIs != nil {
					assert.ErrorIs(t, err, testCase.wantErrIs)
				}
				if testCase.notErrIs != nil {
					assert.NotErrorIs(t, err, testCase.notErrIs)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestCreateVendorBlockFromSpecification(t *testing.T) {
	t.Parallel()

	expectedFirstBlock := VendorBlock{0x22, 0x11, 0x1a, 0x2b, 0xfe, 0x01, 0xdc, 0xba, 0x00, 0x1e}
	expectedSecondBlock := VendorBlock{0x04, 0x43, 0xc3, 0xd4, 0x11, 0x22, 0x33, 0x44, 0xff, 0xff}

	testCases := []struct {
		name          string
		specification VendorSpecification
		expectedBlock *VendorBlock
		wantErr       bool
		wantErrIs     error
	}{
		{
			name: "ValidWeekZero",
			specification: VendorSpecification{
				Manufacturer:      "HPQ",
				ProductCode:       "1A2B",
				SerialNumber:      "FE01DCBA",
				WeekOfManufacture: 0,
				YearOfManufacture: 2020,
			},
			expectedBlock: &expectedFirstBlock,
		},
		{
			name: "ValidWeekSentinel",
			specification: VendorSpecification{
				Manufacturer:      "ABC",
				ProductCode:       "C3D4",
				SerialNumber:      "11223344",
				WeekOfManufacture: 255,
				YearOfManufacture: 2245,
			},
			expectedBlock: &expectedSecondBlock,
		},
		{
			name: "InvalidWeek",
			specification: VendorSpecification{
				Manufacturer:      "HPQ",
				ProductCode:       "1A2B",
				SerialNumber:      "FE01DCBA",
				WeekOfManufacture: 54,
				YearOfManufacture: 2020,
			},
			wantErr:   true,
			wantErrIs: ErrInvalidWeekOfManufacture,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			block, err := CreateVendorBlockFromSpecification(testCase.specification)

			if testCase.wantErr {
				assert.Error(t, err)
				if testCase.wantErrIs != nil {
					assert.ErrorIs(t, err, testCase.wantErrIs)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, block)
			assert.Equal(t, *testCase.expectedBlock, *block)
		})
	}
}

func TestVendorBlockToSpecification(t *testing.T) {
	t.Parallel()

	validWeekZeroBlock := VendorBlock{0x22, 0x11, 0x1a, 0x2b, 0xfe, 0x01, 0xdc, 0xba, 0x00, 0x1e}
	expectedWeekZeroSpecification := VendorSpecification{
		Manufacturer:      "HPQ",
		ProductCode:       "1A2B",
		SerialNumber:      "FE01DCBA",
		WeekOfManufacture: 0,
		YearOfManufacture: 2020,
	}

	validWeekSentinelBlock := VendorBlock{0x04, 0x43, 0xc3, 0xd4, 0x11, 0x22, 0x33, 0x44, 0xff, 0xff}
	expectedWeekSentinelSpecification := VendorSpecification{
		Manufacturer:      "ABC",
		ProductCode:       "C3D4",
		SerialNumber:      "11223344",
		WeekOfManufacture: 255,
		YearOfManufacture: 2245,
	}

	invalidManufacturerBlock := VendorBlock{0x00, 0x00, 0x1a, 0x2b, 0xfe, 0x01, 0xdc, 0xba, 0x00, 0x1e}

	invalidWeekBlock := VendorBlock{0x22, 0x11, 0x1a, 0x2b, 0xfe, 0x01, 0xdc, 0xba, 0x36, 0x1e}

	testCases := []struct {
		name         string
		block        VendorBlock
		expectedSpec *VendorSpecification
		wantErr      bool
		wantErrIs    error
	}{
		{
			name:         "ValidWeekZero",
			block:        validWeekZeroBlock,
			expectedSpec: &expectedWeekZeroSpecification,
		},
		{
			name:         "ValidWeekSentinel",
			block:        validWeekSentinelBlock,
			expectedSpec: &expectedWeekSentinelSpecification,
		},
		{
			name:    "InvalidManufacturer",
			block:   invalidManufacturerBlock,
			wantErr: true,
		},
		{
			name:      "InvalidWeek",
			block:     invalidWeekBlock,
			wantErr:   true,
			wantErrIs: ErrInvalidWeekOfManufacture,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			block := testCase.block
			specification, err := block.ToSpecification()

			if testCase.wantErr {
				assert.Error(t, err)
				if testCase.wantErrIs != nil {
					assert.ErrorIs(t, err, testCase.wantErrIs)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, specification)
			assert.Equal(t, *testCase.expectedSpec, *specification)
		})
	}
}

func TestIsWeekOfManufactureValid(t *testing.T) {
	t.Parallel()

	assert.True(t, isWeekOfManufactureValid(0))
	assert.True(t, isWeekOfManufactureValid(1))
	assert.True(t, isWeekOfManufactureValid(53))
	assert.True(t, isWeekOfManufactureValid(255))
	assert.False(t, isWeekOfManufactureValid(-1))
	assert.False(t, isWeekOfManufactureValid(54))
}
