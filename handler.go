package mongodbz

import (
	"errors"
	"reflect"
)

func NewHandler[T, U any](f func(*MongoClient, T) (U, error)) *Handler[T, U] {
	var tType T
	var uType U
	tt := reflect.TypeOf(tType)
	uu := reflect.TypeOf(uType)
	return &Handler[T, U]{
		TType: FormatType(tt),
		UType: FormatType(uu),
		f:     f,
	}
}

func (h *Handler[T, U]) InputType() string {
	return h.TType
}

func (h *Handler[T, U]) OutputType() string {
	return h.UType
}

func (h *Handler[T, U]) Process(cli *MongoClient, item any, callback func(any, error)) {
	//Check if item can be cast to T, if not, call callback with error
	t, ok := item.(T)
	if !ok {
		callback(nil, errors.New("Item is not of type "+h.TType))
		return
	}
	a, err := h.f(cli, t)
	if callback != nil {
		callback(a, err)
	}
}
func (h *Handler[T, U]) ProcessSync(cli *MongoClient, item any) (any, error) {
	t, ok := item.(T)
	if !ok {
		return nil, errors.New("Item is not of type " + h.TType)
	}
	return h.f(cli, t)
}
