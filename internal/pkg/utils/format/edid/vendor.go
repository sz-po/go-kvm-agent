package edid

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type VendorSpecification struct {
	Manufacturer      string `json:"manufacturer" validate:"required,uppercase,alpha,len=3"`
	ProductCode       string `json:"productCode" validate:"required,uppercase,hexadecimal,len=4"`
	SerialNumber      string `json:"serialNumber" validate:"required,uppercase,hexadecimal,len=8"`
	WeekOfManufacture int    `json:"weekOfManufacture" validate:"min=0,max=255"`
	YearOfManufacture int    `json:"yearOfManufacture" validate:"required,min=1990,max=2245"`
}

var ErrInvalidWeekOfManufacture = errors.New("week of manufacture must be 0, between 1 and 53, or 255")

func (specification *VendorSpecification) Validate() error {
	specificationValidator := validator.New(validator.WithRequiredStructEnabled())

	if err := specificationValidator.Struct(specification); err != nil {
		return err
	}

	if !isWeekOfManufactureValid(specification.WeekOfManufacture) {
		return fmt.Errorf("validate week of manufacture: %w", ErrInvalidWeekOfManufacture)
	}

	return nil
}

type VendorBlock [10]byte

func CreateVendorBlockFromSpecification(specification VendorSpecification) (*VendorBlock, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("specification validation: %w", err)
	}

	block := VendorBlock{}

	if manufacturerBytes, err := EncodeManufacturer(specification.Manufacturer); err != nil {
		return nil, fmt.Errorf("encode manufacturer: %w", err)
	} else {
		copy(block[0:2], manufacturerBytes[:])
	}

	if productCodeBytes, err := hex.DecodeString(specification.ProductCode); err != nil {
		return nil, fmt.Errorf("decode product code: %w", err)
	} else {
		copy(block[2:4], productCodeBytes[:])
	}

	if serialNumberBytes, err := hex.DecodeString(specification.SerialNumber); err != nil {
		return nil, fmt.Errorf("decode serial number: %w", err)
	} else {
		copy(block[4:8], serialNumberBytes[:])
	}

	block[8] = byte(specification.WeekOfManufacture)
	block[9] = byte(specification.YearOfManufacture - 1990)

	return &block, nil
}

func (block *VendorBlock) ToSpecification() (*VendorSpecification, error) {
	specification := &VendorSpecification{}

	if manufacturer, err := DecodeManufacturer([2]byte(block[0:2])); err != nil {
		return nil, fmt.Errorf("decode manufacturer: %w", err)
	} else {
		specification.Manufacturer = manufacturer
	}

	specification.ProductCode = strings.ToUpper(hex.EncodeToString(block[2:4]))
	specification.SerialNumber = strings.ToUpper(hex.EncodeToString(block[4:8]))
	specification.WeekOfManufacture = int(block[8])
	specification.YearOfManufacture = int(block[9]) + 1990

	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("specification validation: %w", err)
	}

	return specification, nil
}

func isWeekOfManufactureValid(week int) bool {
	if week == 0 || week == 255 {
		return true
	}

	return week >= 1 && week <= 53
}
