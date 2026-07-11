package logixclient

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrValidation = errors.New("validation error")
	ErrConnection = errors.New("connection error")
	ErrRequest    = errors.New("request error")
)

type Options struct {
	Address string
	Port    uint
	Path    string
	Timeout time.Duration
	Debug   bool
}

type Identity struct {
	VendorID     uint16 `json:"vendor_id"`
	DeviceType   uint16 `json:"device_type"`
	ProductCode  uint16 `json:"product_code"`
	Revision     string `json:"revision"`
	Status       uint16 `json:"status"`
	SerialNumber uint32 `json:"serial_number"`
	ProductName  string `json:"product_name"`
}

type Program struct {
	Name string `json:"name"`
	ID   uint32 `json:"id"`
}

type Tag struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Instance   uint32 `json:"instance"`
	Dimensions []int  `json:"dimensions,omitempty"`
	Program    string `json:"program,omitempty"`
}

type Client interface {
	Connect() error
	Disconnect() error
	Identity() (Identity, error)
	Programs() ([]Program, error)
	Tags() ([]Tag, error)
	Read(tag string, valueType string, elements uint16) (any, string, error)
	Write(tag string, value any) error
}

type Factory interface {
	New(Options) (Client, error)
}

func ValidateOptions(options Options) error {
	if strings.TrimSpace(options.Address) == "" {
		return fmt.Errorf("%w: address is required", ErrValidation)
	}
	if options.Port == 0 || options.Port > 65535 {
		return fmt.Errorf("%w: port must be between 1 and 65535", ErrValidation)
	}
	if options.Timeout <= 0 {
		return fmt.Errorf("%w: timeout must be positive", ErrValidation)
	}
	return nil
}

func NormaliseType(valueType string) (string, error) {
	valueType = strings.ToLower(strings.TrimSpace(valueType))
	if valueType == "" {
		valueType = "auto"
	}
	switch valueType {
	case "auto", "bool", "sint", "int", "dint", "lint", "usint", "uint", "udint", "ulint", "real", "lreal", "string":
		return valueType, nil
	default:
		return "", fmt.Errorf("%w: unsupported type %q", ErrValidation, valueType)
	}
}

func ParseValue(valueType, raw string) (any, error) {
	valueType, err := NormaliseType(valueType)
	if err != nil {
		return nil, err
	}
	if valueType == "auto" {
		return nil, fmt.Errorf("%w: write requires an explicit --type", ErrValidation)
	}
	parseInt := func(bits int) (int64, error) {
		return strconv.ParseInt(strings.TrimSpace(raw), 0, bits)
	}
	parseUint := func(bits int) (uint64, error) {
		return strconv.ParseUint(strings.TrimSpace(raw), 0, bits)
	}
	var value any
	switch valueType {
	case "bool":
		parsed, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid bool %q", ErrValidation, raw)
		}
		value = parsed
	case "sint":
		parsed, err := parseInt(8)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid sint %q", ErrValidation, raw)
		}
		value = int8(parsed)
	case "int":
		parsed, err := parseInt(16)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid int %q", ErrValidation, raw)
		}
		value = int16(parsed)
	case "dint":
		parsed, err := parseInt(32)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid dint %q", ErrValidation, raw)
		}
		value = int32(parsed)
	case "lint":
		parsed, err := parseInt(64)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid lint %q", ErrValidation, raw)
		}
		value = parsed
	case "usint":
		parsed, err := parseUint(8)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid usint %q", ErrValidation, raw)
		}
		value = uint8(parsed)
	case "uint":
		parsed, err := parseUint(16)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid uint %q", ErrValidation, raw)
		}
		value = uint16(parsed)
	case "udint":
		parsed, err := parseUint(32)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid udint %q", ErrValidation, raw)
		}
		value = uint32(parsed)
	case "ulint":
		parsed, err := parseUint(64)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid ulint %q", ErrValidation, raw)
		}
		value = parsed
	case "real":
		parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 32)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid real %q", ErrValidation, raw)
		}
		value = float32(parsed)
	case "lreal":
		parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid lreal %q", ErrValidation, raw)
		}
		value = parsed
	case "string":
		value = raw
	}
	return value, nil
}
