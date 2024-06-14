package underflow

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hexops/autogold"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type InputTestCase struct {
	want     autogold.Value
	input    interface{}
	typeHint sema.Type
}

var resolver = func(f string, resolverType ResolveType) (string, error) {
	if f == "first" {
		return "0x179b6b1cb6755e31", nil
	}
	return "", nil
}

var (
	stringVal  = "foobar"
	structType = cadence.StructType{
		Fields: []cadence.Field{{
			Identifier: "bar",
			Type:       cadence.StringType,
		}},
	}
	structTypeAddress = cadence.StructType{
		Fields: []cadence.Field{{
			Identifier: "bar",
			Type:       cadence.AddressType,
		}},
	}
)

func TestParseInputValueWithTypeHint(t *testing.T) {
	fix64, _ := cadence.NewUFix64("42.1000000")
	tests := []InputTestCase{
		{
			want:     autogold.Want("string", cadence.String("foo")),
			input:    "foo",
			typeHint: sema.StringType,
		},
		{
			want:     autogold.Want("uint64", cadence.UInt64(42)),
			input:    42,
			typeHint: sema.UInt64Type,
		},
		{
			want:     autogold.Want("uint64 from uint8", cadence.UInt64(42)),
			input:    uint8(42),
			typeHint: sema.UInt64Type,
		},

		{
			want:     autogold.Want("uint32", cadence.UInt32(42)),
			input:    42,
			typeHint: sema.UInt32Type,
		},
		{
			want:     autogold.Want("ufix64", fix64),
			input:    42.1,
			typeHint: sema.UFix64Type,
		},
		{
			want:     autogold.Want("optional string", cadence.Optional{Value: cadence.String("foo")}),
			input:    "foo",
			typeHint: &sema.OptionalType{Type: sema.StringType},
		},
		{
			want:     autogold.Want("optional string empty", cadence.Optional{}),
			input:    nil,
			typeHint: &sema.OptionalType{Type: sema.StringType},
		},
		{
			want:     autogold.Want("address", cadence.Address{23, 155, 107, 28, 182, 117, 94, 49}),
			input:    "first",
			typeHint: sema.TheAddressType,
		},
		{
			want:     autogold.Want("stringPointer", cadence.Optional{Value: cadence.String("foobar")}),
			input:    &stringVal,
			typeHint: &sema.OptionalType{Type: sema.StringType},
		},
		{
			want:     autogold.Want("bool", cadence.Bool(true)),
			input:    true,
			typeHint: sema.BoolType,
		},
		{
			want: autogold.Want("dict string string", cadence.Dictionary{Pairs: []cadence.KeyValuePair{
				{
					Key:   cadence.String("foo"),
					Value: cadence.String("bar"),
				},
			}}),
			input: map[string]string{"foo": "bar"},
			typeHint: &sema.DictionaryType{
				KeyType:   sema.StringType,
				ValueType: sema.StringType,
			},
		},
		{
			want: autogold.Want("slice string", cadence.Array{Values: []cadence.Value{
				cadence.String("foo"),
				cadence.String("bar"),
			}}),
			input: []string{"foo", "bar"},
			typeHint: &sema.VariableSizedType{
				Type: sema.StringType,
			},
		},

		{
			want: autogold.Want("struct", cadence.NewStruct([]cadence.Value{cadence.String("bar")}).WithType(&structType)),
			input: Foo{
				Bar: "bar",
			},
			typeHint: sema.InvalidType, // we will get invalid type here when trying to convert the type
		},
		{
			want: autogold.Want("struct with address", cadence.NewStruct([]cadence.Value{cadence.Address{
				23,
				155,
				107,
				28,
				182,
				117,
				94,
				49,
			}}).WithType(&structTypeAddress),
			),
			input: Debug_Foo2{
				Bar: "0x179b6b1cb6755e31",
			},
			typeHint: sema.InvalidType, // we will get invalid type here when trying to convert the type
		},
		{
			want: autogold.Want("dict string string literal", cadence.Dictionary{
				DictionaryType: &cadence.DictionaryType{
					KeyType:     cadence.PrimitiveType(8),
					ElementType: cadence.PrimitiveType(8),
				},
				Pairs: []cadence.KeyValuePair{
					{
						Key:   cadence.String("test"),
						Value: cadence.String("foo"),
					},
					{
						Key:   cadence.String("test2"),
						Value: cadence.String("bar"),
					},
				},
			}),
			input: `{"test": "foo", "test2": "bar"}`,
			typeHint: &sema.DictionaryType{
				KeyType:   sema.StringType,
				ValueType: sema.StringType,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.want.Name(), func(t *testing.T) {
			val, err := InputToCadenceWithHint(tc.input, tc.typeHint, resolver)
			require.NoError(t, err)
			tc.want.Equal(t, val)
		})
	}
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
		interfaceString,
		int8(8),
		nil,
	}

	for idx, value := range values {
		t.Run(fmt.Sprintf("parse input %d", idx), func(t *testing.T) {
			cv, err := InputToCadence(value, func(string, ResolveType) (string, error) {
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
