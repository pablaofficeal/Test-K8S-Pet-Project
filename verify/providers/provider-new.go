package providers

import "fmt"

func New(name string) (Provider, error) {
	switch name {
	case "test":
		return TestProvider{}, nil
	}
	return nil, fmt.Errorf("unknown verify provider: %s", name)
}