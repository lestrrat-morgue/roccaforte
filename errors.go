package roccaforte

type ignorable interface {
	Ignorable() bool
}

func isIgnorable(err error) bool {
	if i, ok := err.(ignorable); ok {
		return i.Ignorable()
	}
	return false
}
