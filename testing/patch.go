package testing

import (
	"os"
	"reflect"
)

// Restorer holds a function that can be used
// to restore some previous state.
type Restorer func()

// Restore restores some previous state.
func (r Restorer) Restore() {
	r()
}

// PatchValue sets the value pointed to by the given destination to the given
// value, and returns a function to restore it to its original value.  The
// value must be assignable to the element type of the destination.
func PatchValue(dest, value interface{}) Restorer {
	destv := reflect.ValueOf(dest).Elem()
	oldv := reflect.New(destv.Type()).Elem()
	oldv.Set(destv)
	valuev := reflect.ValueOf(value)
	if !valuev.IsValid() {
		// This isn't quite right when the destination type is not
		// nilable, but it's better than the complex alternative.
		valuev = reflect.Zero(destv.Type())
	}
	destv.Set(valuev)
	return func() {
		destv.Set(oldv)
	}
}

// PatchEnvironment provides a test a simple way to override a single
// environment variable. A function is returned that will return the
// environment to what it was before.
func PatchEnvironment(name, value string) Restorer {
	oldValue := os.Getenv(name)
	os.Setenv(name, value)
	return func() {
		os.Setenv(name, oldValue)
	}
}