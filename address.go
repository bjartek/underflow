package underflow

import "github.com/onflow/cadence"

// / This functions extracts out addresses from a cadence value
// /  It currently supports arrays, optionls, dictionaries and structs
func ExtractAddresses(field cadence.Value) []string {
	if field == nil {
		return nil
	}

	switch field := field.(type) {
	case cadence.Optional:
		return ExtractAddresses(field.Value)
	case cadence.Dictionary:
		result := []string{}
		for _, item := range field.Pairs {
			value := ExtractAddresses(item.Value)
			key := getAndUnquoteString(item.Key)

			if value != nil && key != "" {
				result = append(result, value...)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case cadence.Struct:
		result := []string{}
		for _, subField := range field.Fields {
			value := ExtractAddresses(subField)
			if value != nil {
				result = append(result, value...)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case cadence.Array:
		result := []string{}
		for _, item := range field.Values {
			value := ExtractAddresses(item)
			if value != nil {
				result = append(result, value...)
			}
		}
		if len(result) == 0 {
			return nil
		}
		return result
	case cadence.Address:
		return []string{field.String()}
	default:
		return []string{}
	}
}
