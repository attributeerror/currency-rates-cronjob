package main

import "fmt"

type (
	EnvVarNotFoundError struct {
		envVar string
	}

	EnvVarOptions struct {
		Required     bool
		DefaultValue string
	}

	EnvVarOption func(*EnvVarOptions)
)

func (e EnvVarNotFoundError) Error() string {
	return fmt.Sprintf("environment variable not found: %s", e.envVar)
}

func (e EnvVarNotFoundError) Is(template error) bool {
	if template, ok := template.(EnvVarNotFoundError); ok {
		return e.envVar == "" || e.envVar == template.envVar
	}

	return false
}

func WithIsRequired(required bool) EnvVarOption {
	return func(opt *EnvVarOptions) {
		opt.Required = required
	}
}

func WithDefaultValue(defaultVal string) EnvVarOption {
	return func(opt *EnvVarOptions) {
		opt.DefaultValue = defaultVal
	}
}
