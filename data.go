// Copyright (c) 2017 Josh Rickmar
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mill

import (
	"fmt"
	"math"
)

// ValueType describes the type of data stored in a Data.
type ValueType uint

//go:generate stringer -type=ValueType

// Possible types of data stored in a Data.  The zero value represents
const (
	ValueTypeUnknown ValueType = iota
	ValueTypeString
	ValueTypeInt64
	ValueTypeUint64
	ValueTypeFloat64
	ValueTypeAny

	valueTypeMaxValue = ValueTypeAny
)

// Data describes some additional data being logged.  All data is named so it
// can be described and inspected more cleanly using structured logging.
//
// The zero value of this struct is not valid and trying to use it will result
// in panics.  Data structs must only be created using the functions exported by
// this package.
type Data struct {
	name      string
	valueType ValueType
	string    string
	numBits   uint64
	any       interface{}
}

// String returns a Data recording a string.
func String(name string, value string) Data {
	return Data{name: name, valueType: ValueTypeString, string: value}
}

// Int64 returns a Data recording an int64.
func Int64(name string, value int64) Data {
	return Data{name: name, valueType: ValueTypeInt64, numBits: uint64(value)}
}

// Uint64 returns a Data recording a uint64.
func Uint64(name string, value uint64) Data {
	return Data{name: name, valueType: ValueTypeUint64, numBits: value}
}

// Float64 returns a Data recording a float64.
func Float64(name string, value float64) Data {
	return Data{name: name, valueType: ValueTypeFloat64, numBits: math.Float64bits(value)}
}

// Any returns a Data recording any possible value type boxed in an empty
// interface.  Some codecs may
func Any(name string, value interface{}) Data {
	return Data{name: name, valueType: ValueTypeAny, any: value}
}

// Name returns the name of the data field.
func (d *Data) Name() string { return d.name }

// Type returns the type of data described by the Data.
func (d *Data) Type() ValueType { return d.valueType }

func checkType(have, want ValueType) {
	if have != want {
		panic(fmt.Sprintf("value type mismatch: %v != %v", have, want))
	}
}

// String returns the string value contained by the Data.
//
// This function panics if the Data does not describe a string.
func (d *Data) String() string {
	checkType(d.valueType, ValueTypeString)
	return d.string
}

// Int64 returns the int64 value contained by the Data.
//
// This function panics if the Data does not describe an int64.
func (d *Data) Int64() int64 {
	checkType(d.valueType, ValueTypeInt64)
	return int64(d.numBits)
}

// Uint64 returns the uint64 value contained by the Data.
//
// This function panics if the Data does not describe a uint64.
func (d *Data) Uint64() uint64 {
	checkType(d.valueType, ValueTypeUint64)
	return d.numBits
}

// Float64 returns the float64 value contained by the Data.
//
// This function panics if the Data does not describe a float64.
func (d *Data) Float64() float64 {
	checkType(d.valueType, ValueTypeFloat64)
	return math.Float64frombits(d.numBits)
}

// Value returns the value contained by the Data, boxed in an empty interface.
//
// This function panics if the Data is the invalid zero value.
func (d *Data) Value() interface{} {
	switch d.valueType {
	default:
		panic("no data")

	case ValueTypeString:
		return d.string
	case ValueTypeInt64:
		return int64(d.numBits)
	case ValueTypeUint64:
		return d.numBits
	case ValueTypeFloat64:
		return math.Float64frombits(d.numBits)
	case ValueTypeAny:
		switch v := d.any.(type) {
		case fmt.Stringer:
			return v.String()
		default:
			return v
		}
	}
}
