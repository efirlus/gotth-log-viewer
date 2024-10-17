func sprintMap(m interface{}) string {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return "Error: Input is not a map"
	}

	var result strings.Builder
	keys := v.MapKeys()

	// Sort keys if they are comparable
	if len(keys) > 0 && keys[0].Kind() != reflect.Interface {
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
		})
	}

	for _, key := range keys {
		value := v.MapIndex(key)
		result.WriteString(fmt.Sprintf("%v = %v\n", key, value))
	}

	return result.String()
}

func sprintStruct(s interface{}) string {
	var result string
	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		result = result + fmt.Sprintf("%s: %v\n", field.Name, value.Interface())
	}
	return result
}

// Recursive function to find all indexes of a given element
func findIndexes(slice []rune, element rune, startIndex int, result []int) []int {
	// Use slices.Index to find the first occurrence from startIndex
	index := slices.Index(slice[startIndex:], element)

	if index == -1 { // Base case: if no more occurrences, return the result
		return result
	}

	// Adjust the index to be relative to the original slice
	index += startIndex

	// Add the found index to the result
	result = append(result, index)

	// Recursively call findIndexes starting after the found index
	return findIndexes(slice, element, index+1, result)
}

func containsSubsequence(slice, subsequence []rune) bool {
	if len(subsequence) > len(slice) {
		return false
	}
	for i := 0; i <= len(slice)-len(subsequence); i++ {
		if slices.Equal(slice[i:i+len(subsequence)], subsequence) {
			return true
		}
	}
	return false
}

// fillSliceToLength ensures that the slice has at least 'n' length by appending zeros if needed.
func fillSliceToLength(slice []rune, n int) []rune {
	// Check if the slice length is less than the desired length
	if len(slice) < n {
		// Append zeros to the slice until it reaches the desired length
		zeros := make([]int32, n-len(slice)) // Create a slice of zeros
		slice = append(slice, zeros...)      // Append zeros to the original slice
	}
	return slice
}
