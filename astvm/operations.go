package vm

import (
	"github.com/Advik-B/english/runtime"
)

// The arithmetic and comparison operators live in the runtime package, which
// both engines share. These are the names this engine has always exposed; the
// behaviour is defined once, in one place, so the two engines can no longer
// drift apart on what an operator does or how it words a failure.

// Add adds two values: numbers arithmetically, text and arrays by
// concatenation.
func Add(left, right Value) (Value, error) { return runtime.Add(left, right) }

// Subtract subtracts two numbers.
func Subtract(left, right Value) (Value, error) { return runtime.Subtract(left, right) }

// Multiply multiplies two numbers.
func Multiply(left, right Value) (Value, error) { return runtime.Multiply(left, right) }

// Divide divides two numbers. Division by zero is an error.
func Divide(left, right Value) (Value, error) { return runtime.Divide(left, right) }

// Modulo computes the remainder of two numbers.
func Modulo(left, right Value) (Value, error) { return runtime.Modulo(left, right) }

// Compare evaluates a comparison operator and returns a boolean.
func Compare(op string, left, right Value) (bool, error) { return runtime.Compare(op, left, right) }

// Equals reports whether two values are the same type and the same value.
func Equals(left, right Value) bool { return runtime.Equals(left, right) }
