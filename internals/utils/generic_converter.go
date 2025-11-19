package utils

// ConvertToStringMap converts map[interface{}]interface{} to map[string]interface{}
// This is a refactored version using a more generic approach
func ConvertToStringMap(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		key, ok := k.(string)
		if !ok {
			continue
		}
		result[key] = convertValue(v)
	}
	return result
}

// ConvertToStringSlice converts []interface{} to proper string or map slice
func ConvertToStringSlice(a []interface{}) []interface{} {
	result := make([]interface{}, len(a))
	for i, v := range a {
		result[i] = convertValue(v)
	}
	return result
}

// convertValue is a helper function that converts interface{} values recursively
func convertValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		return ConvertToStringMap(val)
	case []interface{}:
		return ConvertToStringSlice(val)
	default:
		return v
	}
}
