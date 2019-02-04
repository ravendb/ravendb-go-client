package ravendb

// RangeBuilder helps build a string query for range requests
type RangeBuilder struct {
	path string

	lessBound        interface{}
	greaterBound     interface{}
	lessInclusive    bool
	greaterInclusive bool

	lessSet    bool
	greaterSet bool

	err error
}

func NewRangeBuilder(path string) *RangeBuilder {
	return &RangeBuilder{
		path: path,
	}
}

func (b *RangeBuilder) createClone() *RangeBuilder {
	builder := *b
	return &builder
}

func (b *RangeBuilder) IsLessThan(value interface{}) *RangeBuilder {
	if b.err != nil {
		return b
	}
	if b.lessSet {
		b.err = newIllegalStateError("Less bound was already set")
		return b
	}

	clone := b.createClone()
	clone.lessBound = value
	clone.lessInclusive = false
	clone.lessSet = true
	return clone
}

func (b *RangeBuilder) IsLessThanOrEqualTo(value interface{}) *RangeBuilder {
	if b.err != nil {
		return b
	}
	if b.lessSet {
		b.err = newIllegalStateError("Less bound was already set")
		return b
	}

	clone := b.createClone()
	clone.lessBound = value
	clone.lessInclusive = true
	clone.lessSet = true
	return clone
}

func (b *RangeBuilder) IsGreaterThan(value interface{}) *RangeBuilder {
	if b.err != nil {
		return b
	}
	if b.greaterSet {
		b.err = newIllegalStateError("Greater bound was already set")
		return b
	}

	clone := b.createClone()
	clone.greaterBound = value
	clone.greaterInclusive = false
	clone.greaterSet = true
	return clone
}

func (b *RangeBuilder) IsGreaterThanOrEqualTo(value interface{}) *RangeBuilder {
	if b.err != nil {
		return b
	}
	if b.greaterSet {
		b.err = newIllegalStateError("Greater bound was already set")
		return b
	}

	clone := b.createClone()
	clone.greaterBound = value
	clone.greaterInclusive = true
	clone.greaterSet = true
	return clone
}

func (b *RangeBuilder) GetStringRepresentation(addQueryParameter func(interface{}) string) (string, error) {
	var less string
	var greater string

	if b.err != nil {
		return "", b.err
	}

	if !b.lessSet && !b.greaterSet {
		return "", newIllegalStateError("Bounds were not set")
	}

	if b.lessSet {
		lessParamName := addQueryParameter(b.lessBound)
		tmp := " < "
		if b.lessInclusive {
			tmp = " <= "
		}
		less = b.path + tmp + "$" + lessParamName
	}

	if b.greaterSet {
		tmp := " > "
		if b.greaterInclusive {
			tmp = " >= "
		}
		greaterParamName := addQueryParameter(b.greaterBound)
		greater = b.path + tmp + "$" + greaterParamName
	}

	if less != "" && greater != "" {
		return greater + " and " + less, nil
	}

	if less != "" {
		return less, nil
	}
	return greater, nil
}

func (b *RangeBuilder) Err() error {
	return b.err
}
