// wrapper for circuit breaker configuration, currently using github.com/afex/hystrix-go

package connect

import "github.com/afex/hystrix-go/hystrix"

// CircuitSetting is used to tune circuit settings at runtime
type CircuitSetting struct {
	// Timeout is how long to wait for command to complete, in milliseconds
	Timeout int `json:"timeout"`

	// MaxConcurrentRequests is how many commands of the same type can run at the same time
	MaxConcurrentRequests int `json:"max_concurrent_requests"`

	// RequestVolumeThreshold is the minimum number of requests needed before a circuit can be tripped due to health
	RequestVolumeThreshold int `json:"request_volume_threshold"`

	// SleepWindow is how long, in milliseconds, to wait after a circuit opens before testing for recovery
	SleepWindow int `json:"sleep_window"`

	// ErrorPercentThreshold causes circuits to open once the rolling measure of errors exceeds this percent of requests
	ErrorPercentThreshold int `json:"error_percent_threshold"`
}

// ConfigureCircuitBreaker is used to set the default value for any method
// hystrix will copy this value as the setting when setting for a command is not found
func ConfigureCircuitBreaker(setting *CircuitSetting) {
	hystrix.DefaultTimeout = setting.Timeout
	hystrix.DefaultMaxConcurrent = setting.MaxConcurrentRequests
	hystrix.DefaultErrorPercentThreshold = setting.ErrorPercentThreshold
	hystrix.DefaultVolumeThreshold = setting.RequestVolumeThreshold
	hystrix.DefaultSleepWindow = setting.SleepWindow
}
