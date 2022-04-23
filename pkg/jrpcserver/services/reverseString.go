package services

// ReverseStringService is an RPC service for string reversal
type ReverseStringService struct {
}

const ReverseStringServiceName string = "reverseString"

// ReverseString reverses a given string using runes
func (service *ReverseStringService) ReverseString(str string) (string, error) {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes), nil
}
