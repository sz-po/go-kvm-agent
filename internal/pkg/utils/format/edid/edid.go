package edid

import "fmt"

type Block [128]byte

const (
	edidBlockLength      = 128
	edidHeaderLength     = 8
	edidVendorLength     = 10
	edidVersionLength    = 2
	edidDisplayLength    = 5
	edidChromacityLength = 10
	edidTimingsLength    = 91
	edidExtensionLength  = 1
	edidChecksumLength   = 1

	edidHeaderStart     = 0
	edidVendorStart     = edidHeaderStart + edidHeaderLength
	edidVersionStart    = edidVendorStart + edidVendorLength
	edidDisplayStart    = edidVersionStart + edidVersionLength
	edidChromacityStart = edidDisplayStart + edidDisplayLength
	edidTimingsStart    = edidChromacityStart + edidChromacityLength
	edidExtensionIndex  = edidTimingsStart + edidTimingsLength
	edidChecksumIndex   = edidExtensionIndex + edidExtensionLength
)

type Specification struct {
	Vendor              VendorSpecification     `json:"vendor" validate:"required"`
	Display             DisplaySpecification    `json:"display" validate:"required"`
	Chromacity          ChromacitySpecification `json:"chromacity" validate:"required"`
	Timings             TimingsSpecification    `json:"timings" validate:"required"`
	ExtensionBlockCount uint8                   `json:"extensionBlockCount" validate:"max=127"`
}

func CreateSpecificationFromBlock(block Block) (*Specification, error) {
	if err := validateEdidChecksum(block); err != nil {
		return nil, fmt.Errorf("checksum: %w", err)
	}

	if err := validateHeader(block[edidHeaderStart : edidHeaderStart+edidHeaderLength]); err != nil {
		return nil, fmt.Errorf("header: %w", err)
	}

	if err := validateVersion(block[edidVersionStart : edidVersionStart+edidVersionLength]); err != nil {
		return nil, fmt.Errorf("version: %w", err)
	}

	specification := &Specification{}

	var vendorBlock VendorBlock
	copy(vendorBlock[:], block[edidVendorStart:edidVendorStart+edidVendorLength])
	if vendorSpecification, err := vendorBlock.ToSpecification(); err != nil {
		return nil, fmt.Errorf("vendor: %w", err)
	} else {
		specification.Vendor = *vendorSpecification
	}

	var displayBlock DisplayBlock
	copy(displayBlock[:], block[edidDisplayStart:edidDisplayStart+edidDisplayLength])
	if displaySpecification, err := CreateDisplaySpecificationFromBlock(displayBlock); err != nil {
		return nil, fmt.Errorf("display: %w", err)
	} else {
		specification.Display = *displaySpecification
	}

	var chromacityBlock ChromacityBlock
	copy(chromacityBlock[:], block[edidChromacityStart:edidChromacityStart+edidChromacityLength])
	if chromacitySpecification, err := CreateChromacitySpecificationFromBlock(chromacityBlock); err != nil {
		return nil, fmt.Errorf("chromacity: %w", err)
	} else {
		specification.Chromacity = *chromacitySpecification
	}

	var timingsBlock TimingsBlock
	copy(timingsBlock[:], block[edidTimingsStart:edidTimingsStart+edidTimingsLength])
	if timingsSpecification, err := CreateTimingsSpecificationFromBlock(timingsBlock); err != nil {
		return nil, fmt.Errorf("timings: %w", err)
	} else {
		specification.Timings = *timingsSpecification
	}

	specification.ExtensionBlockCount = block[edidExtensionIndex]

	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate specification: %w", err)
	}

	return specification, nil
}

func CreateBlockFromSpecification(specification Specification) (*Block, error) {
	if err := specification.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	block := &Block{}

	copy(block[edidHeaderStart:edidHeaderStart+edidHeaderLength], HeaderBlockDefault)

	copy(block[edidVersionStart:edidVersionStart+edidVersionLength], VersionBlockDefault)

	if vendorBlock, err := CreateVendorBlockFromSpecification(specification.Vendor); err != nil {
		return nil, fmt.Errorf("vendor: %w", err)
	} else {
		copy(block[edidVendorStart:edidVendorStart+edidVendorLength], vendorBlock[:])
	}

	if displayBlock, err := CreateDisplayBlockFromSpecification(specification.Display); err != nil {
		return nil, fmt.Errorf("display: %w", err)
	} else {
		copy(block[edidDisplayStart:edidDisplayStart+edidDisplayLength], displayBlock[:])
	}

	if chromacityBlock, err := CreateChromacityBlockFromSpecification(specification.Chromacity); err != nil {
		return nil, fmt.Errorf("chromacity: %w", err)
	} else {
		copy(block[edidChromacityStart:edidChromacityStart+edidChromacityLength], chromacityBlock[:])
	}

	if timingsBlock, err := CreateTimingsBlockFromSpecification(specification.Timings); err != nil {
		return nil, fmt.Errorf("timings: %w", err)
	} else {
		copy(block[edidTimingsStart:edidTimingsStart+edidTimingsLength], timingsBlock[:])
	}

	block[edidExtensionIndex] = specification.ExtensionBlockCount

	block[edidChecksumIndex] = calculateChecksumByte(block[:edidChecksumIndex])

	return block, nil
}

func (specification *Specification) Validate() error {
	if specification == nil {
		return fmt.Errorf("nil specification")
	}

	if specification.ExtensionBlockCount > 127 {
		return fmt.Errorf("invalid extension block count")
	}

	if err := specification.Vendor.Validate(); err != nil {
		return fmt.Errorf("vendor: %w", err)
	}

	if err := specification.Display.Validate(); err != nil {
		return fmt.Errorf("display: %w", err)
	}

	if err := specification.Chromacity.Validate(); err != nil {
		return fmt.Errorf("chromacity: %w", err)
	}

	if err := specification.Timings.Validate(); err != nil {
		return fmt.Errorf("timings: %w", err)
	}

	return nil
}

func validateHeader(headerBytes []byte) error {
	var header HeaderBlock
	copy(header[:], headerBytes)

	if err := header.Validate(); err != nil {
		return err
	}

	return nil
}

func validateVersion(versionBytes []byte) error {
	var versionBlock VersionBlock
	copy(versionBlock[:], versionBytes)

	if err := versionBlock.Validate(); err != nil {
		return err
	}

	return nil
}

func validateEdidChecksum(block Block) error {
	if calculateChecksum(block[:]) != 0 {
		return fmt.Errorf("invalid checksum")
	}

	return nil
}

func calculateChecksum(bytes []byte) byte {
	sum := 0

	for _, value := range bytes {
		sum += int(value)
	}

	return byte(sum % 256)
}

func calculateChecksumByte(prefix []byte) byte {
	sum := 0

	for _, value := range prefix {
		sum += int(value)
	}

	return byte((256 - (sum % 256)) % 256)
}
