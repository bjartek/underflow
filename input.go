package underflow

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structtag"
	"github.com/onflow/cadence"
)

// a resolver to resolve a input type into a name, can be used to resolve struct names for instance
type InputResolver func(string) (string, error)

func InputToCadence(v interface{}, resolver InputResolver) (cadence.Value, error) {
	f := reflect.ValueOf(v)
	return ReflectToCadence(f, resolver)
}

func ReflectToCadence(value reflect.Value, resolver InputResolver) (cadence.Value, error) {
	inputType := value.Type()

	kind := inputType.Kind()
	switch kind {
	case reflect.Interface:
		return cadence.NewValue(value.Interface())
	case reflect.Struct:
		var val []cadence.Value
		fields := []cadence.Field{}
		for i := 0; i < value.NumField(); i++ {
			fieldValue := value.Field(i)
			cadenceVal, err := ReflectToCadence(fieldValue, resolver)
			if err != nil {
				return nil, err
			}
			cadenceType := cadenceVal.Type()

			field := inputType.Field(i)

			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				return nil, err
			}

			name := ""
			tag, err := tags.Get("cadence")
			if err != nil {
				tag, _ = tags.Get("json")
			}
			if tag != nil {
				name = tag.Name
			}

			if name == "-" {
				continue
			}

			if name == "" {
				name = strings.ToLower(field.Name)
			}

			if IsTagCadecenAddress(tag) {
				stringVal := getAndUnquoteString(cadenceVal)
				adr, err := hexToAddress(stringVal)
				if err != nil {
					return nil, err
				}
				cadenceAddress := cadence.BytesToAddress(adr.Bytes())
				cadenceType = cadence.AddressType
				cadenceVal = cadenceAddress
			}

			fields = append(fields, cadence.Field{
				Identifier: name,
				Type:       cadenceType,
			})

			val = append(val, cadenceVal)
		}

		resolvedIdentifier, err := resolver(inputType.Name())
		if err != nil {
			return nil, err
		}
		structType := cadence.StructType{
			QualifiedIdentifier: resolvedIdentifier,
			Fields:              fields,
		}

		structValue := cadence.NewStruct(val).WithType(&structType)
		return structValue, nil

	case reflect.Pointer:
		if value.IsNil() {
			return cadence.NewOptional(nil), nil
		}

		ptrValue, err := ReflectToCadence(value.Elem(), resolver)
		if err != nil {
			return nil, err
		}
		return cadence.NewOptional(ptrValue), nil

	case reflect.Int:
		return cadence.NewInt(value.Interface().(int)), nil
	case reflect.Int8:
		return cadence.NewInt8(value.Interface().(int8)), nil
	case reflect.Int16:
		return cadence.NewInt16(value.Interface().(int16)), nil
	case reflect.Int32:
		return cadence.NewInt32(value.Interface().(int32)), nil
	case reflect.Int64:
		return cadence.NewInt64(value.Interface().(int64)), nil
	case reflect.Bool:
		return cadence.NewBool(value.Interface().(bool)), nil
	case reflect.Uint:
		return cadence.NewUInt(value.Interface().(uint)), nil
	case reflect.Uint8:
		return cadence.NewUInt8(value.Interface().(uint8)), nil
	case reflect.Uint16:
		return cadence.NewUInt16(value.Interface().(uint16)), nil
	case reflect.Uint32:
		return cadence.NewUInt32(value.Interface().(uint32)), nil
	case reflect.Uint64:
		return cadence.NewUInt64(value.Interface().(uint64)), nil
	case reflect.String:
		result, err := cadence.NewString(value.Interface().(string))
		return result, err
	case reflect.Float64:
		result, err := cadence.NewUFix64(fmt.Sprintf("%f", value.Interface().(float64)))
		return result, err

	case reflect.Map:
		array := []cadence.KeyValuePair{}
		iter := value.MapRange()

		for iter.Next() {
			key := iter.Key()
			val := iter.Value()
			cadenceKey, err := ReflectToCadence(key, resolver)
			if err != nil {
				return nil, err
			}
			cadenceVal, err := ReflectToCadence(val, resolver)
			if err != nil {
				return nil, err
			}
			array = append(array, cadence.KeyValuePair{Key: cadenceKey, Value: cadenceVal})
		}
		return cadence.NewDictionary(array), nil
	case reflect.Slice, reflect.Array:
		array := []cadence.Value{}
		for i := 0; i < value.Len(); i++ {
			arrValue := value.Index(i)
			cadenceVal, err := ReflectToCadence(arrValue, resolver)
			if err != nil {
				return nil, err
			}
			array = append(array, cadenceVal)
		}
		return cadence.NewArray(array), nil

	}

	return nil, fmt.Errorf("Not supported type for now. Type : %s", inputType.Kind())
}

func IsTagCadecenAddress(tag *structtag.Tag) bool {
	if tag == nil {
		return false
	}

	for _, opt := range tag.Options {
		if opt == "cadenceAddress" {
			return true
		}
	}
	return false
}
