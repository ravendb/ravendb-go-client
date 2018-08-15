package ravendb

type RangeBuilder struct {
	path string

	lessBound        interface{}
	greaterBound     interface{}
	lessInclusive    bool
	greaterInclusive bool

	lessSet    bool
	greaterSet bool
}

func NewRangeBuilder(path string) *RangeBuilder {
	return &RangeBuilder{
		path: path,
	}
}

func RangeBuilder_forPath(path string) *RangeBuilder {
	return NewRangeBuilder(path)
}

func (b *RangeBuilder) CreateClone() *RangeBuilder {
	builder := *b
	return &builder
}

func (b *RangeBuilder) IsLessThan(value interface{}) *RangeBuilder {
	if b.lessSet {
		//throw new IllegalStateException("Less bound was already set")
		panic("Less bound was already set")
	}

	clone := b.CreateClone()
	clone.lessBound = value
	clone.lessInclusive = false
	clone.lessSet = true
	return clone
}

func (b *RangeBuilder) IsLessThanOrEqualTo(value interface{}) *RangeBuilder {
	if b.lessSet {
		//throw new IllegalStateException("Less bound was already set")
		panic("Less bound was already set")
	}

	clone := b.CreateClone()
	clone.lessBound = value
	clone.lessInclusive = true
	clone.lessSet = true
	return clone
}

func (b *RangeBuilder) IsGreaterThan(value interface{}) *RangeBuilder {
	if b.greaterSet {
		//throw new IllegalStateException("Greater bound was already set")
		panic("Greater bound was already set")
	}

	clone := b.CreateClone()
	clone.greaterBound = value
	clone.greaterInclusive = false
	clone.greaterSet = true
	return clone
}

func (b *RangeBuilder) IsGreaterThanOrEqualTo(value interface{}) *RangeBuilder {
	if b.greaterSet {
		//throw new IllegalStateException("Greater bound was already set")
		panic("Greater bound was already set")
	}

	clone := b.CreateClone()
	clone.greaterBound = value
	clone.greaterInclusive = true
	clone.greaterSet = true
	return clone
}

func (b *RangeBuilder) GetStringRepresentation(addQueryParameter func(Object) string) string {
	var less string
	var greater string

	if !b.lessSet && !b.greaterSet {
		//throw new IllegalStateException("Bounds were not set")
		panic("Bounds were not set")
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
		return greater + " and " + less
	}

	if less != "" {
		return less
	}
	return greater
}
