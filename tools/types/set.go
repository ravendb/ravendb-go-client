package types


//todo: adding new type - analog SET Python later
type ISet interface {
	Add(interface{})
	Delete(interface{})
	Append(ISet)
	HasKey(interface{})
	Interset(ISet)
}
type Set map[interface{}] struct{}
func (ref Set) Add(newKey interface{})  {
	ref[newKey] = struct{}{}
}
func (ref Set) Delete(key interface{})  {
	delete(ref,key)
}
func (ref Set) HasKey(key interface{}) bool {
	_, ok := ref[key]

	return ok
}
func (ref Set) Append(newSet Set)  {
	for key, _ := range newSet {
		ref.Add(key)
	}
}
func NewSETFromArray(array []string) (*Set){
	set := make(Set, len(array))
	for _, val := range array{
		set.Add(val)
	}
	return &set
}

type SETstr map[string] struct{}

func (ref SETstr) Add(newKey string)  {
		ref[newKey] = struct{}{}
}
func (ref SETstr) Delete(key string)  {
	delete(ref,key)
}
func (ref SETstr) HasKey(key string) bool {
	_, ok := ref[key]

	return ok
}
func (ref SETstr) Append(newSet Set)  {
	for key, _ := range newSet {
		ref.Add(key.(string))
	}
}
func (ref SETstr) ToSlice() []string {
	res := make([]string, 0, len(ref))
	for key := range ref {
		res = append(res, key)
	}

	return res
}
func (ref SETstr) Clear() {
	ref = make(SETstr, 0)
}
type SETint map[int] struct{}

func (ref SETint) Add(newKey int)  {
	ref[newKey] = struct{}{}
}
func (ref SETint) Delete(key int)  {
	delete(ref,key)
}
func (ref SETint) HasKey(key int) bool {
	_, ok := ref[key]

	return ok
}
func (ref SETint) Append(newSet Set)  {
	for key, _ := range newSet {
		ref.Add(key.(int))
	}
}

func NewSETstrFromArray(array []string) (*SETstr){
	setStr := make(SETstr, len(array))
	for _, val := range array{
		setStr.Add(val)
	}
	return &setStr
}

func NewSETintFromArray(array []int) (*SETint){
	setInt := make(SETint, len(array))
	for _, val := range array{
		setInt.Add(val)
	}
	return &setInt
}
