package elog

import "elog/internal/core"

type D = core.Field

// KVString construct Field with string value.
func KVString(key string, value string) D {
	return D{Key: key, Type: core.StringType, StringVal: value}
}

func KV(key string, value interface{}) D {
	return D{Key: key, Value: value}
}