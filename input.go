package underflow

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structtag"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/pkg/errors"
)

type ResolveType int

const (
	Address ResolveType = iota
	Identifier
)

func (d ResolveType) String() string {
	return [...]string{"Address", "Identiifer"}[d]
}

// a resolver to resolve a input type into a name, can be used to resolve struct names for instance
type InputResolver func(string, ResolveType) (string, error)

var flowInterpeter, _ = interpreter.NewInterpreter(nil, nil, &interpreter.Config{})

func InputToCadenceWithHint(v interface{}, typeHint sema.Type, resolver InputResolver) (cadence.Value, error) {
	cadenceVal, isCadenceValue := v.(cadence.Value)
	if isCadenceValue {
		return cadenceVal, nil
	}
	f := reflect.ValueOf(v)
	return ReflectToCadenceWithTypeHint(f, typeHint, resolver)
}

func ReflectToCadenceWithTypeHint(value reflect.Value, typeHint sema.Type, resolver InputResolver) (cadence.Value, error) {
	if value == reflect.ValueOf(nil) {
		return cadence.NewOptional(nil), nil
	}
	inputType := value.Type()

	kind := inputType.Kind()
	switch kind {
	case reflect.Struct:
		var val []cadence.Value
		fields := []cadence.Field{}
		for i := 0; i < value.NumField(); i++ {
			fieldValue := value.Field(i)
			// We do not have type hint here so we have to just revert to the method without it
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

		resolvedIdentifier, err := resolver(inputType.Name(), Identifier)
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
		hint := typeHint.(*sema.OptionalType)
		ptrValue, err := ReflectToCadenceWithTypeHint(value.Elem(), hint.Type, resolver)
		if err != nil {
			return nil, errors.Wrap(err, "getting value of optional type")
		}
		return cadence.NewOptional(ptrValue), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return runtime.ParseLiteral(fmt.Sprintf("%v", value), typeHint, flowInterpeter)
	case reflect.Bool:
		return cadence.NewBool(value.Interface().(bool)), nil
	case reflect.String:
		stringVal := value.Interface().(string)
		th := typeHint
		optionalHint, isOptionalType := typeHint.(*sema.OptionalType)
		if isOptionalType {
			th = optionalHint.Type
		}
		if th == sema.TheAddressType {
			result, err := resolver(stringVal, Address)
			if err != nil {
				return nil, err
			}
			adr, err := hexToAddress(result)
			if err != nil {
				return nil, err
			}
			cadenceAddress := cadence.BytesToAddress(adr.Bytes())
			return cadenceAddress, nil
		} else if th == sema.StringType {
			if len(stringVal) > 0 && !strings.HasPrefix(stringVal, "\"") {
				stringVal = "\"" + stringVal + "\""
			}
		}

		return runtime.ParseLiteral(stringVal, typeHint, flowInterpeter)
	case reflect.Float64:
		arg := fmt.Sprintf("%f", value.Interface().(float64))
		return runtime.ParseLiteral(arg, typeHint, flowInterpeter)

	case reflect.Map:
		array := []cadence.KeyValuePair{}
		iter := value.MapRange()
		hint := typeHint.(*sema.DictionaryType)
		for iter.Next() {
			key := iter.Key()
			val := iter.Value()
			cadenceKey, err := ReflectToCadenceWithTypeHint(key, hint.KeyType, resolver)
			if err != nil {
				return nil, err
			}
			cadenceVal, err := ReflectToCadenceWithTypeHint(val, hint.ValueType, resolver)
			if err != nil {
				return nil, err
			}
			array = append(array, cadence.KeyValuePair{Key: cadenceKey, Value: cadenceVal})
		}
		return cadence.NewDictionary(array), nil
	case reflect.Slice, reflect.Array:
		array := []cadence.Value{}

		arrayType := typeHint.(sema.ArrayType)
		for i := 0; i < value.Len(); i++ {
			arrValue := value.Index(i)
			cadenceVal, err := ReflectToCadenceWithTypeHint(arrValue, arrayType.ElementType(false), resolver)
			if err != nil {
				return nil, err
			}
			array = append(array, cadenceVal)
		}
		return cadence.NewArray(array), nil

	}

	return nil, fmt.Errorf("not supported type for now. Type : %s", inputType.Kind())
}

func InputToCadence(v interface{}, resolver InputResolver) (cadence.Value, error) {
	f := reflect.ValueOf(v)
	return ReflectToCadence(f, resolver)
}

func NewCadenceValue(value any) (cadence.Value, error) {
	switch v := value.(type) {
	case string:
		return cadence.NewString(v)
	case int:
		return cadence.NewInt(v), nil
	case int8:
		return cadence.NewInt8(v), nil
	case int16:
		return cadence.NewInt16(v), nil
	case int32:
		return cadence.NewInt32(v), nil
	case int64:
		return cadence.NewInt64(v), nil
	case uint8:
		return cadence.NewUInt8(v), nil
	case uint16:
		return cadence.NewUInt16(v), nil
	case uint32:
		return cadence.NewUInt32(v), nil
	case uint64:
		return cadence.NewUInt64(v), nil
	case []any:
		values := make([]cadence.Value, len(v))

		for i, v := range v {
			t, err := NewCadenceValue(v)
			if err != nil {
				return nil, err
			}

			values[i] = t
		}

		return cadence.NewArray(values), nil
	case nil:
		return cadence.NewOptional(nil), nil
	}

	return nil, fmt.Errorf("value type %T cannot be converted to ABI value type", value)
}

func ReflectToCadence(value reflect.Value, resolver InputResolver) (cadence.Value, error) {
	if value == reflect.ValueOf(nil) {
		return cadence.NewOptional(nil), nil
	}
	inputType := value.Type()

	kind := inputType.Kind()
	switch kind {
	case reflect.Interface:
		return NewCadenceValue(value.Interface())
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

		resolvedIdentifier, err := resolver(inputType.Name(), Identifier)
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

	return nil, fmt.Errorf("not supported type for now. Type : %s", inputType.Kind())
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
