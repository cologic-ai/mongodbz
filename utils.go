package mongodbz

import (
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	"reflect"
)

func getInt(val interface{}) int {
	switch val.(type) {
	case int:
		return val.(int)
	case int32:
		return int(val.(int32))
	case int64:
		return int(val.(int64))
	default:
		return 0
	}
}

func FormatType(t reflect.Type) string {
	if t == nil {
		return "nil"
	}
	return t.PkgPath() + "." + t.Name()
}
