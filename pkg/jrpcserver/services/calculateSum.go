package services

// CalculateSumService is an RPC service for sum calculation
type CalculateSumService struct {
}

const CalculateSumServiceName string = "calculateSum"

// CalculateSum computes sum of two given floats
func (service *CalculateSumService) CalculateSum(a, b float64) (float64, error) {
	return a + b, nil
}
