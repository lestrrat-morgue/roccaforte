package incoming

func (r Rule) Disabled() bool {
	return false
}

func (r Rule) AggregationWindow() int64 {
	return 300
}