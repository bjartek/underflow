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

// / Convert a cadence value into a interface{} structure for easier consumption in go with options
func CadenceValueToInterfaceWithOption(field cadence.Value, opt Options) interface{} {
	if field == nil {
		return nil
	}

	switch field := field.(type) {
	case cadence.Optional:
		return CadenceValueToInterfaceWithOption(field.Value, opt)
	case cadence.Dictionary:
		// fmt.Println("is dict ", field.ToGoValue(), " ", field.String())
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
		// fmt.Println("is struct ", field.ToGoValue(), " ", field.String())
		result := map[string]interface{}{}
		subStructNames := field.StructType.Fields

		for j, subField := range field.Fields {
			value := CadenceValueToInterfaceWithOption(subField, opt)
			key := subStructNames[j].Identifier

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
		// fmt.Println("is array ", field.ToGoValue(), " ", field.String())
		var result []interface{}
		for _, item := range field.Values {
			value := CadenceValueToInterfaceWithOption(item, opt)
			//	fmt.Printf("%+v\n", value)
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
	case cadence.UInt:
		return field.String()
	case cadence.Address:
		return field.String()
	case cadence.TypeValue:
		// fmt.Println("is type ", field.ToGoValue(), " ", field.String())
		return field.StaticType.ID()
	case cadence.String:
		// fmt.Println("is string ", field.ToGoValue(), " ", field.String())
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

		for i, subField := range field.Fields {
			value := CadenceValueToInterfaceWithOption(subField, opt)
			if value != nil || opt.IncludeEmptyValues {
				result[field.EventType.Fields[i].Identifier] = value
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
		subStructNames := field.ResourceType.Fields

		for j, subField := range field.Fields {
			value := CadenceValueToInterfaceWithOption(subField, opt)
			key := subStructNames[j].Identifier

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
	default:
		// fmt.Println("is fallthrough ", field.ToGoValue(), " ", field.String())

		goValue := field.ToGoValue()
		if goValue != nil {
			return goValue
		}
		return field.String()
	}
}
