package main

import (
	"syscall/js"
)

var (
	document = js.Global().Get("document")
	window   = js.Global().Get("window")
	location = js.Global().Get("location")
	THREE    = js.Global().Get("THREE")
	THREEx   = js.Global().Get("THREEx")
)

type goObject struct {
	jsValue js.Value
}

func (g goObject) ref() js.Value {
	return g.jsValue
}

type Promise[T any] interface {
	Then(cb func(value T)) Promise[T]
	Catch(cb func(err error)) Promise[T]
}

var _ Promise[struct{}] = goPromise[struct{}]{}

type goPromise[T any] struct {
	goObject
	convert func(value js.Value) T
}

func (g goPromise[T]) Then(cb func(value T)) Promise[T] {
	var jsFunc js.Func
	jsFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		defer jsFunc.Release()
		cb(g.convert(args[0]))
		return nil
	})
	g.jsValue.Call("then", jsFunc)
	return g
}

func (g goPromise[T]) Catch(cb func(err error)) Promise[T] {
	var jsFunc js.Func
	jsFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		defer jsFunc.Release()
		cb(js.Error{
			Value: args[0],
		})
		return js.Undefined()
	})
	g.jsValue.Call("catch", jsFunc)
	return g
}

func Import(url string) Promise[js.Value] {
	return goPromise[js.Value]{
		goObject: goObject{jsValue: js.Global().Call("import", url)},
		convert: func(value js.Value) js.Value {
			return value
		},
	}
}
