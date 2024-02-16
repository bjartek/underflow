package underflow

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/hexops/autogold"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CadenceTest struct {
	want  autogold.Value
	input cadence.Value
}

func TestCadenceValueToInterface(t *testing.T) {
	foo := cadenceString("foo")
	bar := cadenceString("bar")
	emptyString := cadenceString("")

	emptyStrct := cadence.Struct{
		Fields: []cadence.Value{emptyString},
		StructType: &cadence.StructType{
			Fields: []cadence.Field{{
				Identifier: "foo",
				Type:       cadence.StringType,
			}},
		},
	}

	address1, _ := hex.DecodeString("f8d6e0586b0a20c7")
	caddress1, _ := common.BytesToAddress(address1)
	structType := cadence.StructType{
		Location:            common.NewAddressLocation(nil, caddress1, ""),
		QualifiedIdentifier: "Contract.Bar",
		Fields: []cadence.Field{{
			Identifier: "foo",
			Type:       cadence.StringType,
		}},
	}
	strct := cadence.Struct{
		Fields:     []cadence.Value{bar},
		StructType: &structType,
	}
	dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: foo, Value: bar}})

	emoji := cadenceString("😁")
	emojiDict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: emoji, Value: emoji}})

	cadenceAddress1 := cadence.BytesToAddress(address1)

	cadenceEvent := cadence.NewEvent([]cadence.Value{foo}).WithType(&cadence.EventType{
		QualifiedIdentifier: "TestEvent",
		Fields: []cadence.Field{{
			Type:       cadence.StringType,
			Identifier: "foo",
		}},
	},
	)

	resource := cadence.Resource{
		ResourceType: &cadence.ResourceType{
			Location:            common.NewAddressLocation(nil, caddress1, ""),
			QualifiedIdentifier: "Contract.Resource",
			Fields: []cadence.Field{{
				Identifier: "foo",
				Type:       cadence.StringType,
			}},
		},
		Fields: []cadence.Value{foo},
	}

	stringType := cadence.StringType
	path := cadence.Path{Domain: common.PathDomainStorage, Identifier: "foo"}
	pathCap := cadence.NewCapability(1, cadenceAddress1, cadence.StringType)

	structTypeValue := cadence.NewTypeValue(&structType)
	stringTypeValue := cadence.NewTypeValue(&stringType)
	ufix, _ := cadence.NewUFix64("42.0")
	fix, _ := cadence.NewFix64("-2.0")

	largeUfix, _ := cadence.NewUFix64("184467440737.0")
	smallfix, _ := cadence.NewFix64("-92233720368.5")
	largefix, _ := cadence.NewFix64("92233720368.5")
	var ui64 uint64 = math.MaxUint64

	testCases := []CadenceTest{
		{autogold.Want("EmptyString", nil), cadenceString("")},
		{autogold.Want("nil", nil), nil},
		{autogold.Want("None", nil), cadence.NewOptional(nil)},
		{autogold.Want("Some(string)", "foo"), cadence.NewOptional(foo)},
		{autogold.Want("Some(uint64)", uint64(42)), cadence.NewOptional(cadence.NewUInt64(42))},
		{autogold.Want("uint64", uint64(42)), cadence.NewUInt64(42)},
		{autogold.Want("max uint64", ui64), cadence.NewUInt64(ui64)},
		{autogold.Want("ufix64", float64(42.0)), ufix},
		{autogold.Want("large_ufix64", float64(1.84467440737e+11)), largeUfix},
		{autogold.Want("fix64", float64(-2.0)), fix},
		{autogold.Want("small_fix64", float64(-9.22337203685e+10)), smallfix},
		{autogold.Want("large_fix64", float64(9.22337203685e+10)), largefix},
		{autogold.Want("uint32", uint32(42)), cadence.NewUInt32(42)},
		{autogold.Want("int", 42), cadence.NewInt(42)},
		{autogold.Want("string array", []interface{}{"foo", "bar"}), cadence.NewArray([]cadence.Value{foo, bar})},
		{autogold.Want("empty array", nil), cadence.NewArray([]cadence.Value{emptyString})},
		{autogold.Want("string array ignore empty", []interface{}{"foo", "bar"}), cadence.NewArray([]cadence.Value{foo, emptyString, bar})},
		{autogold.Want("dictionary", map[string]interface{}{"foo": "bar"}), dict},
		{autogold.Want("dictionary_ignore_empty_value", nil), cadence.NewDictionary([]cadence.KeyValuePair{{Key: foo, Value: emptyString}})},
		{autogold.Want("dictionary_with_subdict", map[string]interface{}{"bar": map[string]interface{}{"foo": "bar"}}), cadence.NewDictionary([]cadence.KeyValuePair{{Key: bar, Value: dict}})},
		{autogold.Want("struct", map[string]interface{}{"foo": "bar"}), strct},
		{autogold.Want("empty struct", nil), emptyStrct},
		{autogold.Want("address", "0xf8d6e0586b0a20c7"), cadenceAddress1},
		{autogold.Want("string type", "String"), stringTypeValue},
		{autogold.Want("struct type", "A.f8d6e0586b0a20c7.Contract.Bar"), structTypeValue},
		{autogold.Want("Emoji", "😁"), emoji},
		{autogold.Want("EmojiDict", map[string]interface{}{"😁": "😁"}), emojiDict},
		{autogold.Want("StoragePath", "/storage/foo"), path},
		{autogold.Want("Event", map[string]interface{}{"foo": "foo"}), cadenceEvent},
		{autogold.Want("PathCapablity", map[string]interface{}{"address": "0xf8d6e0586b0a20c7", "id": 1}), pathCap},
		{autogold.Want("Resource", map[string]interface{}{"foo": "foo"}), resource},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			value := CadenceValueToInterface(tc.input)
			tc.want.Equal(t, value)
		})
	}
}

func TestCadenceValueToJson(t *testing.T) {
	result, err := CadenceValueToJsonString(cadence.String(""))
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestParseInputValue(t *testing.T) {
	foo := "foo"

	var interfaceString interface{} = "foo"
	var strPointer *string = nil
	values := []interface{}{
		"foo",
		uint64(42),
		map[string]uint64{"foo": uint64(42)},
		[]uint64{42, 69},
		[2]string{"foo", "bar"},
		&foo,
		strPointer,
		float64(2.0),
		uint(1.0),
		interfaceString,
		int8(8),
	}

	for idx, value := range values {
		t.Run(fmt.Sprintf("parse input %d", idx), func(t *testing.T) {
			cv, err := InputToCadence(value, func(string) (string, error) {
				return "", nil
			})
			assert.NoError(t, err)
			v := CadenceValueToInterface(cv)

			vj, err := json.Marshal(v)
			assert.NoError(t, err)

			cvj, err := json.Marshal(value)
			assert.NoError(t, err)

			assert.Equal(t, string(cvj), string(vj))
		})
	}
}

func TestMarshalCadenceStruct(t *testing.T) {
	val, err := InputToCadence(Foo{Bar: "foo"}, func(string) (string, error) {
		return "A.123.Foo.Bar", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "A.123.Foo.Bar", val.Type().ID())
	jsonVal, err := CadenceValueToJsonString(val)
	assert.NoError(t, err)
	assert.JSONEq(t, `{ "bar": "foo" }`, jsonVal)
}

func TestMarshalCadenceStructWithStructTag(t *testing.T) {
	val, err := InputToCadence(Foo{Bar: "foo"}, func(string) (string, error) {
		return "A.123.Foo.Baz", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "A.123.Foo.Baz", val.Type().ID())
	jsonVal, err := CadenceValueToJsonString(val)
	assert.NoError(t, err)
	assert.JSONEq(t, `{ "bar": "foo" }`, jsonVal)
}

// TODO: this might actually need an integration test to be useful
func TestMarshalCadenceStructWithAddressStructTag(t *testing.T) {
	val, err := InputToCadence(Debug_Foo2{Bar: "0xf8d6e0586b0a20c7"}, func(string) (string, error) {
		return "A.123.Debug.Foo2", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "A.123.Debug.Foo2", val.Type().ID())
	jsonVal, err := CadenceValueToJsonString(val)
	assert.NoError(t, err)
	assert.JSONEq(t, `{ "bar": "0xf8d6e0586b0a20c7" }`, jsonVal)
}

func TestPrimitiveInputToCadence(t *testing.T) {
	tests := []struct {
		value interface{}
		name  string
	}{
		{name: "int", value: 1},
		{name: "int8", value: int8(8)},
		{name: "int16", value: int16(16)},
		{name: "int32", value: int32(32)},
		{name: "int64", value: int64(64)},
		{name: "uint8", value: uint8(8)},
		{name: "uint16", value: uint16(16)},
		{name: "uint32", value: uint32(32)},
		{name: "true", value: true},
		{name: "false", value: false},
	}

	resolver := func(string) (string, error) {
		return "", nil
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cadenceValue, err := InputToCadence(test.value, resolver)
			assert.NoError(t, err)
			result2 := CadenceValueToInterface(cadenceValue)
			assert.Equal(t, test.value, result2)
		})
	}
}

func TestExtractAddresses(t *testing.T) {
	address1, err := hexToAddress("f8d6e0586b0a20c7")
	require.NoError(t, err)
	address := *address1

	address2Ptr, err := hexToAddress("01cf0e2f2f715450")
	require.NoError(t, err)
	address2 := *address2Ptr

	opt := cadence.Optional{
		Value: address,
	}

	dict := cadence.NewDictionary([]cadence.KeyValuePair{
		{Key: cadenceString("owner"), Value: address},
		{Key: cadenceString("sender"), Value: address2},
	})

	array := cadence.NewArray([]cadence.Value{address, address2})

	structType := cadence.StructType{
		QualifiedIdentifier: "Contract.Bar",
		Fields: []cadence.Field{{
			Identifier: "owner",
			Type:       cadence.AddressType,
		}},
	}
	strct := cadence.Struct{
		Fields:     []cadence.Value{address},
		StructType: &structType,
	}
	testCases := []CadenceTest{
		{autogold.Want("Address", []string{"0xf8d6e0586b0a20c7"}), address},
		{autogold.Want("OptAddress", []string{"0xf8d6e0586b0a20c7"}), opt},
		{autogold.Want("Dict", []string{"0xf8d6e0586b0a20c7", "0x01cf0e2f2f715450"}), dict},
		{autogold.Want("Array", []string{"0xf8d6e0586b0a20c7", "0x01cf0e2f2f715450"}), array},
		{autogold.Want("Struct", []string{"0xf8d6e0586b0a20c7"}), strct},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			value := ExtractAddresses(tc.input)
			tc.want.Equal(t, value)
		})
	}
}

func TestIncludeEmptyValues(t *testing.T) {
	dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: cadenceString("foo"), Value: cadenceString("")}})
	array := cadence.NewArray([]cadence.Value{cadenceString("foo"), cadenceString(""), cadenceString("bar")})
	structType := cadence.StructType{
		QualifiedIdentifier: "Contract.Bar",
		Fields: []cadence.Field{{
			Identifier: "foo",
			Type:       cadence.StringType,
		}},
	}

	strct := cadence.Struct{
		Fields:     []cadence.Value{cadenceString("")},
		StructType: &structType,
	}

	cadenceEvent := cadence.NewEvent([]cadence.Value{cadenceString("")}).WithType(&cadence.EventType{
		QualifiedIdentifier: "TestEvent",
		Fields: []cadence.Field{{
			Type:       cadence.StringType,
			Identifier: "foo",
		}},
	},
	)

	testCases := []CadenceTest{
		{autogold.Want("Dict", map[string]interface{}{"foo": ""}), dict},
		{autogold.Want("Array", []interface{}{"foo", "", "bar"}), array},
		{autogold.Want("Struct", map[string]interface{}{"foo": ""}), strct},
		{autogold.Want("Event", map[string]interface{}{"foo": ""}), cadenceEvent},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			value := CadenceValueToInterfaceWithOption(tc.input, Options{
				IncludeEmptyValues: true,
			})
			tc.want.Equal(t, value)
		})
	}
}

func TestUseStringsForFixedNumbers(t *testing.T) {
	ufix, _ := cadence.NewUFix64("42.0")
	fix, _ := cadence.NewFix64("-2.0")

	largeUfix, _ := cadence.NewUFix64("184467440737.0")
	smallfix, _ := cadence.NewFix64("-92233720368.5")
	largefix, _ := cadence.NewFix64("92233720368.5")

	testCases := []CadenceTest{
		{autogold.Want("ufix64", "42.00000000"), ufix},
		{autogold.Want("large_ufix64", "184467440737.00000000"), largeUfix},
		{autogold.Want("fix64", "-2.00000000"), fix},
		{autogold.Want("small_fix64", "-92233720368.50000000"), smallfix},
		{autogold.Want("large_fix64", "92233720368.50000000"), largefix},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			value := CadenceValueToInterfaceWithOption(tc.input, Options{
				UseStringForFixedNumbers: true,
			})
			tc.want.Equal(t, value)
		})
	}
}

func TestWrapWithComplextTypes(t *testing.T) {
	address1, _ := hex.DecodeString("f8d6e0586b0a20c7")
	caddress1, _ := common.BytesToAddress(address1)
	structType := cadence.StructType{
		Location:            common.NewAddressLocation(nil, caddress1, ""),
		QualifiedIdentifier: "Contract.Bar",
		Fields: []cadence.Field{{
			Identifier: "foo",
			Type:       cadence.StringType,
		}},
	}

	strct := cadence.Struct{
		Fields:     []cadence.Value{cadenceString("Foo")},
		StructType: &structType,
	}

	cadenceEvent := cadence.NewEvent([]cadence.Value{cadenceString("Foo")}).WithType(&cadence.EventType{
		Location:            common.NewAddressLocation(nil, caddress1, ""),
		QualifiedIdentifier: "Contract.TestEvent",
		Fields: []cadence.Field{{
			Type:       cadence.StringType,
			Identifier: "foo",
		}},
	},
	)

	stringType := cadence.StringType

	resource := cadence.Resource{
		ResourceType: &cadence.ResourceType{
			Location:            common.NewAddressLocation(nil, caddress1, ""),
			QualifiedIdentifier: "Contract.Resource",
			Fields: []cadence.Field{{
				Identifier: "foo",
				Type:       cadence.StringType,
			}},
		},
		Fields: []cadence.Value{cadenceString("foo")},
	}

	cadenceAddress1 := cadence.BytesToAddress(address1)
	pathCap := cadence.NewCapability(1, cadenceAddress1, stringType)

	testCases := []CadenceTest{
		{autogold.Want("Struct", map[string]interface{}{"<A.f8d6e0586b0a20c7.Contract.Bar>": map[string]interface{}{"foo": "Foo"}}), strct},
		{autogold.Want("Event", map[string]interface{}{"<A.f8d6e0586b0a20c7.Contract.TestEvent>": map[string]interface{}{"foo": "Foo"}}), cadenceEvent},
		{autogold.Want("Resource", map[string]interface{}{"<@A.f8d6e0586b0a20c7.Contract.Resource>": map[string]interface{}{"foo": "foo"}}), resource},
		{autogold.Want("PathCap", map[string]interface{}{"<Capability<String>>": map[string]interface{}{"address": "0xf8d6e0586b0a20c7", "id": 1}}), pathCap},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			value := CadenceValueToInterfaceWithOption(tc.input, Options{
				WrapWithComplexTypes: true,
			})
			tc.want.Equal(t, value)
		})
	}
}

// in Debug.cdc
type Foo struct {
	Bar string
}

type Debug_FooListBar struct {
	Bar string
	Foo []Debug_Foo2
}

type Debug_FooBar struct {
	Bar string
	Foo Debug_Foo
}

type Debug_Foo_Skip struct {
	Bar  string
	Skip string `cadence:"-"`
}

type Debug_Foo2 struct {
	Bar string `cadence:"bar,cadenceAddress"`
}

type Debug_Foo struct {
	Bar string
}

// in Foo.Bar.Baz
type Baz struct {
	Something string `json:"bar"`
}
