package poly

import "github.com/nofeaturesonlybugs/set"

// SlicesTypeList allows slices of built in primitives to be targets when
// unmarshaling forms, path params, and queries.
var SlicesTypeList = set.NewTypeList(
	[]bool{},
	[]float32{}, []float64{},
	[]int{}, []int8{}, []int16{}, []int32{}, []int64{},
	[]uint{}, []uint8{}, []uint16{}, []uint32{}, []uint64{},
	[]string{},
)

// DefaultFormMapper is a *set.Mapper instance with reasonable defaults for mapping incoming
// *http.Request form data to destination structs.
var DefaultFormMapper = &set.Mapper{
	Tags:             []string{"form"},
	TaggedFieldsOnly: true,
	TreatAsScalar:    SlicesTypeList,
}

// DefaultPathMapper is a *set.Mapper instance with reasonable defaults for mapping incoming
// *http.Request URI path data to destination structs.
var DefaultPathMapper = &set.Mapper{
	Tags:             []string{"path"},
	TaggedFieldsOnly: true,
	TreatAsScalar:    SlicesTypeList,
}

// DefaultQueryMapper is a *set.Mapper instance with reasonable defaults for mapping incoming
// *http.Request query string data to destination structs.
var DefaultQueryMapper = &set.Mapper{
	Tags:             []string{"query"},
	TaggedFieldsOnly: true,
	TreatAsScalar:    SlicesTypeList,
}
