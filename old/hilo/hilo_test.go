// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hilo

import (
	"testing"
	"time"
)

// TestNewHiLoReturnCommand_EmptyParam must error test if not get error during Create ReturnCommand with empty parameters
func TestNewHiLoReturnCommand_EmptyParam(t *testing.T) {
	_, err := NewHiLoReturnCommand("", 0, 0)
	if err == nil {
		t.Error("Not return error message during call NewNextHiLoCommand with empty parameter TAG")
		return
	}
}
func TestNewHiLoReturnCommand(t *testing.T) {
	ref, err := NewHiLoReturnCommand("test", 1, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if ref.Method != "PUT" {
		t.Error("not correct Method property")
	}
}
func TestNewNextHiLoCommand_EmptyParams(t *testing.T) {
	_, err := NewNextHiLoCommand("", 0, time.Time{}, "", 0)
	if err == nil {
		t.Error("Not return error message during call NewNextHiLoCommand with empty string parameters")
		return
	}
}
func TestNewNextHiLoCommand(t *testing.T) {

	ref, err := NewNextHiLoCommand("test", 0, time.Time{}, "test", 0)
	if err != nil {
		t.Errorf(err.Error())
	}
	if ref.Method != "GET" {
		t.Error("not correct Method property")
	}
}
func TestNewHiLoKeyGenerator(t *testing.T) {
	root := NewMultiDatabaseHiLoKeyGenerator("test", "localhost", nil)
	parent := NewMultiTypeHiLoKeyGenerator(*root)
	ref := NewHiLoKeyGenerator("test", *parent)
	if (ref.rangeValues.min_id != 1) || (ref.rangeValues.max_id != 0) {
		t.Error("not valid property RangeValues")
	}

}
