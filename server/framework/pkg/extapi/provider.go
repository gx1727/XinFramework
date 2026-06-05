package extapi

var globalProvider Provider

func Get() Provider {
	if globalProvider == nil {
		panic("extapi Provider is not initialized")
	}
	return globalProvider
}

func Set(p Provider) {
	globalProvider = p
}
