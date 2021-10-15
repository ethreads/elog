package core

// FieldType represent D value type
type FieldType uint32

// D Type enum
const (
	UnknownType FieldType = iota
	StringType
	IntType
	Int64Type
	UintType
	Uint64Type
	Float32Type
	Float64Type
	DurationType
)

// Field is for encoder
type Field struct {
	Key string
	Value interface{}
	Type      FieldType
	StringVal string
	Int64Val  int64
}

