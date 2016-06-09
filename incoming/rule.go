package incoming

import (
	"github.com/pkg/errors"
)

type ignorableErr struct {
	error
}

func (e ignorableErr) Ignorable() bool {
	return true
}

func wrapIgnorable(err error) error {
	return ignorableErr{error: err}
}

func (m *RuleMap) Get(name string) (*Rule, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.rules == nil {
		return nil, wrapIgnorable(errors.New("rule not found: '" + name + "'"))
	}

	r, ok := m.rules[name]
	if !ok {
		return nil, wrapIgnorable(errors.New("rule not found: '" + name + "'"))
	}
	return r, nil
}

func (m *RuleMap) Set(name string, r *Rule) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.rules == nil {
		m.rules = make(map[string]*Rule)
	}

	m.rules[name] = r
}

func (r Rule) Disabled() bool {
	return false
}

func (r Rule) AggregationWindow() int64 {
	return 300
}