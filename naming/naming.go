package naming

import "errors"

// errors
var (
	ErrNotFound = errors.New("service no found")
)

// Naming defined methods of the naming service
type Naming interface {
	Find(serviceName string) ([]ServiceRegistration, error)
	Remove(serviceName, serviceID string) error
	Register(ServiceRegistration) error
	Deregister(serviceID string) error
}
