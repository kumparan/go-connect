// wrapper for circuit breaker configuration, currently using github.com/afex/hystrix-go

package connect

import "github.com/afex/hystrix-go/hystrix"

// CircuitSetting is used to tune circuit settings at runtime
type CircuitSetting struct {
	Timeout                int `json:"timeout"`
	MaxConcurrentRequests  int `json:"max_concurrent_requests"`
	RequestVolumeThreshold int `json:"request_volume_threshold"`
	SleepWindow            int `json:"sleep_window"`
	ErrorPercentThreshold  int `json:"error_percent_threshold"`
}

// ConfigureCircuitBreaker is used to set the default value for any method
// hystrix will copy this value as the setting when setting for a command is not found
func ConfigureCircuitBreaker(setting *CircuitSetting) {
	hystrix.ConfigureCommand("any", hystrix.CommandConfig{
		Timeout:                setting.Timeout,
		MaxConcurrentRequests:  setting.MaxConcurrentRequests,
		ErrorPercentThreshold:  setting.ErrorPercentThreshold,
		RequestVolumeThreshold: setting.RequestVolumeThreshold,
		SleepWindow:            setting.SleepWindow,
	})
}
