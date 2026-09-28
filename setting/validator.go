package setting

import (
	"errors"
	"fmt"
)

// ConfigValidator is implemented by configuration sections that can check
// themselves.
type ConfigValidator interface {
	Validate() error
	GetWarnings() []string
}

// Validate runs the full static validation of the service definition and
// returns nil when the configuration may be applied. All problems are
// aggregated; use ValidateProxyService to get the structured findings.
func (ps *ProxyService) Validate() error {
	if ps == nil {
		return errors.New("proxy service is nil")
	}
	return ValidateProxyService(ps).Error()
}

// GetWarnings returns non-fatal findings of the service configuration.
func (ps *ProxyService) GetWarnings() []string {
	if ps == nil {
		return nil
	}
	rep := ValidateProxyService(ps)
	if len(rep.Warnings) == 0 {
		return nil
	}
	out := make([]string, 0, len(rep.Warnings))
	for _, w := range rep.Warnings {
		out = append(out, fmt.Sprintf("%s: %s", ps.Id, w.String()))
	}
	return out
}
