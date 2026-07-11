package logixclient

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/danomagnum/gologix"
)

type GoLogixFactory struct{}

func (GoLogixFactory) New(options Options) (Client, error) {
	if err := ValidateOptions(options); err != nil {
		return nil, err
	}
	client := gologix.NewClient(options.Address)
	client.Controller.Port = options.Port
	client.SocketTimeout = options.Timeout
	if options.Debug {
		client.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	if options.Path == "" {
		client.Controller.Path = &bytes.Buffer{}
	} else {
		path, err := gologix.ParsePath(options.Path)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid CIP path %q: %v", ErrValidation, options.Path, err)
		}
		client.Controller.Path = path
	}
	return &goLogixClient{client: client}, nil
}

type goLogixClient struct {
	client *gologix.Client
}

func (c *goLogixClient) Connect() error {
	if err := c.client.Connect(); err != nil {
		return fmt.Errorf("%w: %v", ErrConnection, err)
	}
	return nil
}

func (c *goLogixClient) Disconnect() error {
	if err := c.client.Disconnect(); err != nil {
		return fmt.Errorf("%w: disconnect: %v", ErrConnection, err)
	}
	return nil
}

func (c *goLogixClient) Identity() (Identity, error) {
	readUint16 := func(attribute gologix.CIPAttribute) (uint16, error) {
		item, err := c.client.GetAttrSingle(gologix.CipObject_Identity, gologix.CIPInstance(1), attribute)
		if err != nil {
			return 0, err
		}
		return item.Uint16()
	}
	readUint32 := func(attribute gologix.CIPAttribute) (uint32, error) {
		item, err := c.client.GetAttrSingle(gologix.CipObject_Identity, gologix.CIPInstance(1), attribute)
		if err != nil {
			return 0, err
		}
		return item.Uint32()
	}

	vendorID, err := readUint16(gologix.CIPAttribute(1))
	if err != nil {
		return Identity{}, identityError("vendor id", err)
	}
	deviceType, err := readUint16(gologix.CIPAttribute(2))
	if err != nil {
		return Identity{}, identityError("device type", err)
	}
	productCode, err := readUint16(gologix.CIPAttribute(3))
	if err != nil {
		return Identity{}, identityError("product code", err)
	}
	revisionItem, err := c.client.GetAttrSingle(gologix.CipObject_Identity, gologix.CIPInstance(1), gologix.CIPAttribute(4))
	if err != nil {
		return Identity{}, identityError("revision", err)
	}
	major, err := revisionItem.Byte()
	if err != nil {
		return Identity{}, identityError("revision major", err)
	}
	minor, err := revisionItem.Byte()
	if err != nil {
		return Identity{}, identityError("revision minor", err)
	}
	status, err := readUint16(gologix.CIPAttribute(5))
	if err != nil {
		return Identity{}, identityError("status", err)
	}
	serial, err := readUint32(gologix.CIPAttribute(6))
	if err != nil {
		return Identity{}, identityError("serial number", err)
	}
	nameItem, err := c.client.GetAttrSingle(gologix.CipObject_Identity, gologix.CIPInstance(1), gologix.CIPAttribute(7))
	if err != nil {
		return Identity{}, identityError("product name", err)
	}
	nameLength, err := nameItem.Byte()
	if err != nil {
		return Identity{}, identityError("product name length", err)
	}
	if len(nameItem.Rest()) < int(nameLength) {
		return Identity{}, identityError("product name", fmt.Errorf("response declared %d bytes but contained %d", nameLength, len(nameItem.Rest())))
	}

	return Identity{
		VendorID:     vendorID,
		DeviceType:   deviceType,
		ProductCode:  productCode,
		Revision:     fmt.Sprintf("%d.%d", major, minor),
		Status:       status,
		SerialNumber: serial,
		ProductName:  string(nameItem.Rest()[:nameLength]),
	}, nil
}

func identityError(field string, err error) error {
	return fmt.Errorf("%w: read identity %s: %v", ErrRequest, field, err)
}

func (c *goLogixClient) Programs() ([]Program, error) {
	if err := c.client.ListAllPrograms(); err != nil {
		return nil, fmt.Errorf("%w: list programs: %v", ErrRequest, err)
	}
	programs := make([]Program, 0, len(c.client.KnownPrograms))
	for _, program := range c.client.KnownPrograms {
		programs = append(programs, Program{Name: program.Name, ID: uint32(program.ID)})
	}
	sort.Slice(programs, func(i, j int) bool { return programs[i].Name < programs[j].Name })
	return programs, nil
}

func (c *goLogixClient) Tags() ([]Tag, error) {
	if err := c.client.ListAllTags(0); err != nil {
		return nil, fmt.Errorf("%w: list tags: %v", ErrRequest, err)
	}
	tags := make([]Tag, 0, len(c.client.KnownTags))
	for _, known := range c.client.KnownTags {
		program := ""
		if known.Parent != nil {
			program = known.Parent.Name
		}
		dimensions := append([]int(nil), known.Array_Order...)
		tags = append(tags, Tag{
			Name:       known.Name,
			Type:       known.Info.Type.String(),
			Instance:   uint32(known.Instance),
			Dimensions: dimensions,
			Program:    program,
		})
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i].Name < tags[j].Name })
	return tags, nil
}

func (c *goLogixClient) Read(tag string, valueType string, elements uint16) (any, string, error) {
	valueType, err := NormaliseType(valueType)
	if err != nil {
		return nil, "", err
	}
	cipType, err := cipTypeFor(valueType)
	if err != nil {
		return nil, "", err
	}
	value, err := c.client.Read_single(tag, cipType, elements)
	if err != nil {
		return nil, "", fmt.Errorf("%w: read %s: %v", ErrRequest, tag, err)
	}
	if valueType == "string" {
		value = normaliseStringValue(value)
	}
	actualType := valueType
	if actualType == "auto" {
		actualType = inferValueType(value)
	}
	return value, actualType, nil
}

func (c *goLogixClient) Write(tag string, value any) error {
	if err := c.client.Write(tag, value); err != nil {
		return fmt.Errorf("%w: write %s: %v", ErrRequest, tag, err)
	}
	return nil
}

func cipTypeFor(valueType string) (gologix.CIPType, error) {
	switch valueType {
	case "auto":
		return gologix.CIPTypeUnknown, nil
	case "bool":
		return gologix.CIPTypeBOOL, nil
	case "sint":
		return gologix.CIPTypeSINT, nil
	case "int":
		return gologix.CIPTypeINT, nil
	case "dint":
		return gologix.CIPTypeDINT, nil
	case "lint":
		return gologix.CIPTypeLINT, nil
	case "usint":
		return gologix.CIPTypeUSINT, nil
	case "uint":
		return gologix.CIPTypeUINT, nil
	case "udint":
		return gologix.CIPTypeUDINT, nil
	case "ulint":
		return gologix.CIPTypeULINT, nil
	case "real":
		return gologix.CIPTypeREAL, nil
	case "lreal":
		return gologix.CIPTypeLREAL, nil
	case "string":
		return gologix.CIPTypeSTRING, nil
	default:
		return gologix.CIPTypeUnknown, fmt.Errorf("%w: unsupported type %q", ErrValidation, valueType)
	}
}

func inferValueType(value any) string {
	if value == nil {
		return "unknown"
	}
	t := reflect.TypeOf(value)
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int8:
		return "sint"
	case reflect.Int16:
		return "int"
	case reflect.Int32:
		return "dint"
	case reflect.Int64:
		return "lint"
	case reflect.Uint8:
		return "usint"
	case reflect.Uint16:
		return "uint"
	case reflect.Uint32:
		return "udint"
	case reflect.Uint64:
		return "ulint"
	case reflect.Float32:
		return "real"
	case reflect.Float64:
		return "lreal"
	case reflect.String:
		return "string"
	default:
		return strings.ToLower(t.String())
	}
}

func normaliseStringValue(value any) any {
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	case []any:
		values := make([]string, len(typed))
		for i, item := range typed {
			switch item := item.(type) {
			case []byte:
				values[i] = string(item)
			case string:
				values[i] = item
			default:
				values[i] = fmt.Sprint(item)
			}
		}
		return values
	default:
		return value
	}
}
