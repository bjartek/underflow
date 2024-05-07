package underflow

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/onflow/cadence"
)

type Options struct {
	IncludeEmptyValues       bool
	WrapWithComplexTypes     bool
	UseStringForFixedNumbers bool
}

var defaultOptions = Options{
	IncludeEmptyValues:       false,
	WrapWithComplexTypes:     false,
	UseStringForFixedNumbers: false,
}

// / This method converts a cadence.Value to an json string representing that value
func CadenceValueToJsonString(value cadence.Value) (string, error) {
	return CadenceValueToJsonStringWithOption(value, defaultOptions)
}

// / This method converts a cadence.Value to an json string representing that value using the sendt in options to control how it is done
func CadenceValueToJsonStringWithOption(value cadence.Value, opt Options) (string, error) {
	result := CadenceValueToInterfaceWithOption(value, opt)
	if result == nil {
		return "", nil
	}
	j, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return "", err
	}

	return string(j), nil
}

// / Convert a cadence value into a interface{} structure for easier consumption in go
func CadenceValueToInterface(field cadence.Value) interface{} {
	return CadenceValueToInterfaceWithOption(field, defaultOptions)
}

/*
*
values.go:var _ Value = Function{}
*/
// / Convert a cadence value into a interface{} structure for easier consumption in go with options
func CadenceValueToInterfaceWithOption(field cadence.Value, opt Options) interface{} {
	if field == nil {
		return nil
	}

	switch field := field.(type) {
	case cadence.Optional:
		return CadenceValueToInterfaceWithOption(field.Value, opt)
	case cadence.Dictionary:
		result := map[string]interface{}{}
		for _, item := range field.Pairs {
			value := CadenceValueToInterfaceWithOption(item.Value, opt)
			key := getAndUnquoteString(item.Key)

			if key != "" {
				if value != nil || opt.IncludeEmptyValues {
					result[key] = value
				}
			}
		}

		if len(result) == 0 && !opt.IncludeEmptyValues {
			return nil
		}
		return result
	case cadence.Struct:
		result := map[string]interface{}{}
		subFields := cadence.FieldsMappedByName(field)
		for key, subField := range subFields {
			value := CadenceValueToInterfaceWithOption(subField, opt)

			//	fmt.Println("struct ", key, "value", value)
			if value != nil || opt.IncludeEmptyValues {
				result[key] = value
			}
		}
		if len(result) == 0 && !opt.IncludeEmptyValues {
			return nil
		}

		if !opt.WrapWithComplexTypes {
			return result
		}

		return map[string]interface{}{
			fmt.Sprintf("<%s>", field.StructType.ID()): result,
		}
	case cadence.Array:
		var result []interface{}
		for _, item := range field.Values {
			value := CadenceValueToInterfaceWithOption(item, opt)
			if value != nil || opt.IncludeEmptyValues {
				result = append(result, value)
			}
		}
		if len(result) == 0 && !opt.IncludeEmptyValues {
			return nil
		}
		return result

	case cadence.Int:
		return field.String()

	case cadence.Int8:
		return int8(field)
	case cadence.Int16:
		return int16(field)
	case cadence.Int32:
		return int32(field)
	case cadence.Int64:
		return int64(field)
	case cadence.Int128:
		return field.String()
	case cadence.Int256:
		return field.String()
	case cadence.UInt8:
		return uint8(field)
	case cadence.UInt16:
		return uint16(field)
	case cadence.UInt32:
		return uint32(field)
	case cadence.UInt64:
		return uint64(field)
	case cadence.UInt128:
		return field.String()
	case cadence.UInt256:
		return field.String()
	case cadence.Word8:
		return uint8(field)
	case cadence.Word16:
		return uint16(field)
	case cadence.Word32:
		return uint32(field)
	case cadence.Word64:
		return uint64(field)
	case cadence.Word128:
		return field.String()
	case cadence.Word256:
		return field.String()
	case cadence.UInt:
		return field.String()
	case cadence.Address:
		return field.String()
	case *cadence.InclusiveRange:
		return field.String()
	case cadence.TypeValue:
		return field.StaticType.ID()
	case cadence.String:
		value := getAndUnquoteString(field)
		if value == "" && !opt.IncludeEmptyValues {
			return nil
		}
		return value

	case cadence.UFix64:
		if opt.UseStringForFixedNumbers {
			return field.String()
		}
		// fmt.Println("is ufix64 ", field.ToGoValue(), " ", field.String())

		float, _ := strconv.ParseFloat(field.String(), 64)
		return float
	case cadence.Fix64:
		if opt.UseStringForFixedNumbers {
			return field.String()
		}
		float, _ := strconv.ParseFloat(field.String(), 64)
		return float
	case cadence.Event:
		result := map[string]interface{}{}

		subFields := cadence.FieldsMappedByName(field)
		for key, subField := range subFields {
			value := CadenceValueToInterfaceWithOption(subField, opt)
			if value != nil || opt.IncludeEmptyValues {
				result[key] = value
			}
		}

		if !opt.WrapWithComplexTypes {
			return result
		}

		return map[string]interface{}{
			fmt.Sprintf("<%s>", field.EventType.ID()): result,
		}

	case cadence.Resource:

		fields := map[string]interface{}{}
		// fmt.Println("is struct ", field.ToGoValue(), " ", field.String())
		subFields := cadence.FieldsMappedByName(field)
		for key, subField := range subFields {
			value := CadenceValueToInterfaceWithOption(subField, opt)

			//	fmt.Println("struct ", key, "value", value)
			if value != nil || opt.IncludeEmptyValues {
				fields[key] = value
			}
		}

		if !opt.WrapWithComplexTypes {
			return fields
		}

		return map[string]interface{}{
			fmt.Sprintf("<@%s>", field.ResourceType.ID()): fields,
		}

	case cadence.Attachment:

		fields := map[string]interface{}{}
		// fmt.Println("is struct ", field.ToGoValue(), " ", field.String())
		subFields := cadence.FieldsMappedByName(field)
		for key, subField := range subFields {
			value := CadenceValueToInterfaceWithOption(subField, opt)

			//	fmt.Println("struct ", key, "value", value)
			if value != nil || opt.IncludeEmptyValues {
				fields[key] = value
			}
		}

		if !opt.WrapWithComplexTypes {
			return fields
		}

		return map[string]interface{}{
			fmt.Sprintf("<Attachment<%s>>", field.AttachmentType.ID()): fields,
		}

	case cadence.Contract:
		fields := map[string]interface{}{}
		subFields := cadence.FieldsMappedByName(field)
		for key, subField := range subFields {
			value := CadenceValueToInterfaceWithOption(subField, opt)

			//	fmt.Println("struct ", key, "value", value)
			if value != nil || opt.IncludeEmptyValues {
				fields[key] = value
			}
		}

		if !opt.WrapWithComplexTypes {
			return fields
		}

		return map[string]interface{}{
			fmt.Sprintf("<Contract<%s>>", field.ContractType.ID()): fields,
		}

	case cadence.Capability:

		fields := map[string]interface{}{
			"address": CadenceValueToInterfaceWithOption(field.Address, opt),
			"id":      CadenceValueToInterfaceWithOption(field.ID, opt),
		}
		if !opt.WrapWithComplexTypes {
			return fields
		}
		return map[string]interface{}{
			fmt.Sprintf("<Capability<%s>>", field.BorrowType.ID()): fields,
		}
	case cadence.Bool:
		return bool(field)
	case cadence.Bytes:
		return []byte(field)
	case cadence.Character:
		return string(field)
	case cadence.Path:
		return field.String()
	case *cadence.TypeValue:
		return field.String()
	case cadence.Enum:
		return field.String()
	case cadence.Function:
		return field.FunctionType.ID()
	default:
		return field.String()
	}
}
