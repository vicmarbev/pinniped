// Copyright 2020-2022 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package testlogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sort"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

// newStdLogger returns a logr.Logger that matches the legacy v0.4.0 stdr.New implementation.
// All unnecessary functionality has been stripped out.  Avoid using this if possible.
func newStdLogger(std stdr.StdLogger) logr.Logger {
	return logr.New(logger{
		std:    std,
		prefix: "",
		values: nil,
	})
}

type logger struct {
	std    stdr.StdLogger
	prefix string
	values []interface{}
}

func (l logger) clone() logger {
	out := l
	l.values = copySlice(l.values)
	return out
}

func copySlice(in []interface{}) []interface{} {
	out := make([]interface{}, len(in))
	copy(out, in)
	return out
}

// Magic string for intermediate frames that we should ignore.
const autogeneratedFrameName = "<autogenerated>"

// Discover how many frames we need to climb to find the caller. This approach
// was suggested by Ian Lance Taylor of the Go team, so it *should* be safe
// enough (famous last words).
func framesToCaller() int {
	// 1 is the immediate caller.  3 should be too many.
	for i := 1; i < 3; i++ {
		_, file, _, _ := runtime.Caller(i + 1) // +1 for this function's frame
		if file != autogeneratedFrameName {
			return i
		}
	}
	return 1 // something went wrong, this is safe
}

func flatten(kvList ...interface{}) string {
	keys := make([]string, 0, len(kvList))
	vals := make(map[string]interface{}, len(kvList))
	for i := 0; i < len(kvList); i += 2 {
		k, ok := kvList[i].(string)
		if !ok {
			panic(fmt.Sprintf("key is not a string: %s", pretty(kvList[i])))
		}
		var v interface{}
		if i+1 < len(kvList) {
			v = kvList[i+1]
		}
		keys = append(keys, k)
		vals[k] = v
	}
	sort.Strings(keys)
	buf := bytes.Buffer{}
	for i, k := range keys {
		v := vals[k]
		if i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(pretty(k))
		buf.WriteString("=")
		buf.WriteString(pretty(v))
	}
	return buf.String()
}

func pretty(value interface{}) string {
	jb, _ := json.Marshal(value)
	return string(jb)
}

func (l logger) Info(level int, msg string, kvList ...interface{}) {
	if l.Enabled(level) {
		builtin := make([]interface{}, 0, 4)
		builtin = append(builtin, "level", level, "msg", msg)
		builtinStr := flatten(builtin...)
		fixedStr := flatten(l.values...)
		userStr := flatten(kvList...)
		l.output(framesToCaller(), fmt.Sprintln(l.prefix, builtinStr, fixedStr, userStr))
	}
}

func (l logger) Enabled(level int) bool {
	return true
}

func (l logger) Error(err error, msg string, kvList ...interface{}) {
	builtin := make([]interface{}, 0, 4)
	builtin = append(builtin, "msg", msg)
	builtinStr := flatten(builtin...)
	var loggableErr interface{}
	if err != nil {
		loggableErr = err.Error()
	}
	errStr := flatten("error", loggableErr)
	fixedStr := flatten(l.values...)
	userStr := flatten(kvList...)
	l.output(framesToCaller(), fmt.Sprintln(l.prefix, builtinStr, errStr, fixedStr, userStr))
}

func (l logger) output(calldepth int, s string) {
	depth := calldepth + 2 // offset for this adapter

	// ignore errors - what can we really do about them?
	if l.std != nil {
		_ = l.std.Output(depth, s)
	} else {
		_ = log.Output(depth, s)
	}
}

func (l logger) V(level int) logr.LogSink {
	return l.clone()
}

// WithName returns a new logr.Logger with the specified name appended.  stdr
// uses '/' characters to separate name elements.  Callers should not pass '/'
// in the provided name string, but this library does not actually enforce that.
func (l logger) WithName(name string) logr.LogSink {
	new := l.clone()
	if len(l.prefix) > 0 {
		new.prefix = l.prefix + "/"
	}
	new.prefix += name
	return new
}

// WithValues returns a new logr.Logger with the specified key-and-values
// saved.
func (l logger) WithValues(kvList ...interface{}) logr.LogSink {
	new := l.clone()
	new.values = append(new.values, kvList...)
	return new
}

func (l logger) WithCallDepth(depth int) logr.LogSink {
	return l.clone()
}

var _ logr.LogSink = logger{}
var _ logr.CallDepthLogSink = logger{}

func (l logger) Init(info logr.RuntimeInfo) {}