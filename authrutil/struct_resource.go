package authrutil

import (
	"reflect"
	"regexp"

	"github.com/cloudflare/authr"
)

var exportedField = regexp.MustCompile("^[A-Z]")

type structResource struct {
	typ string
	v   reflect.Value
}

func (s structResource) GetResourceType() (string, error) {
	return s.typ, nil
}

func (s structResource) GetResourceAttribute(key string) (interface{}, error) {
	if !exportedField.MatchString(key) {
		return nil, nil
	}
	f, ok := s.v.Type().FieldByName(key)
	if !ok {
		return nil, nil
	}
	return s.v.FieldByIndex(f.Index).Interface(), nil
}

var _ authr.Resource = structResource{}

// StructResource accepts a string that indicates the "rsrc_type" of a resource,
// and the struct that needs to be acceptable as an authr.Resource. This
// function will panic if v is NOT a struct.
func StructResource(typ string, v interface{}) authr.Resource {
	if reflect.TypeOf(v).Kind() != reflect.Struct {
		panic("authrutil.StructResource provided with a non-struct value")
	}
	return structResource{typ: typ, v: reflect.ValueOf(v)}
}
