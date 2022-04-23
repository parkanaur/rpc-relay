package services

// ServiceRegistry is a collection of all services referred to by their names
var ServiceRegistry = map[string]func() interface{}{
	CalculateSumServiceName:  func() interface{} { return &CalculateSumService{} },
	ReverseStringServiceName: func() interface{} { return &ReverseStringService{} },
}
