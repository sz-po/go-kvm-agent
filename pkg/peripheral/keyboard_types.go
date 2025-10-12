package peripheral

// KeyboardHIDUsage represents a HID usage identifier for keyboard keys.
type KeyboardHIDUsage uint16

// KeyboardLogicalKey captures the logical meaning of a key within a layout.
type KeyboardLogicalKey struct {
	Code        string
	Description string
}

// KeyboardLayoutBinding maps a HID usage to logical keys under various modifier states.
// It defines how a physical key produces different logical outputs depending on
// active modifiers (Shift, AltGr, CapsLock) and provides metadata about the key's behavior.
type KeyboardLayoutBinding struct {
	Primary      KeyboardLogicalKey
	Shifted      KeyboardLogicalKey
	AltGr        KeyboardLogicalKey
	ShiftAltGr   KeyboardLogicalKey
	CapsLock     KeyboardLogicalKey
	Alternative  []KeyboardLogicalKey
	IsDeadKey    bool
	ProducesRune bool
	ProducesText bool
}

// KeyboardLayout describes a keyboard layout and the logical bindings it provides.
type KeyboardLayout struct {
	ID          string
	Variant     string
	Description string
	Locale      string
	Bindings    map[KeyboardHIDUsage]KeyboardLayoutBinding
}

// KeyboardInfo contains information about a keyboard device.
type KeyboardInfo struct {
	// Manufacturer identifies the keyboard vendor (e.g., Logitech).
	Manufacturer string

	// Model contains the human-readable product name.
	Model string

	// SerialNumber carries the device-unique identifier when available.
	SerialNumber string

	// SupportedLayouts enumerates layouts the device or firmware exposes for negotiation.
	SupportedLayouts []KeyboardLayout

	// CurrentLayout describes the layout the source is currently applying to key events.
	CurrentLayout KeyboardLayout

	// MaxSimultaneousKeys caps the number of concurrent key presses the device can emit.
	MaxSimultaneousKeys uint32

	// FirmwareRevision records the version string reported by the keyboard firmware.
	FirmwareRevision string

	// PhysicalKeyOverrides remaps specific HID usages to layout bindings for this device only.
	PhysicalKeyOverrides map[KeyboardHIDUsage]KeyboardLayoutBinding
}

// KeyboardLEDState represents the state of keyboard indicator LEDs.
type KeyboardLEDState struct {
	CapsLock   bool
	NumLock    bool
	ScrollLock bool
	Custom     map[string]bool
}
