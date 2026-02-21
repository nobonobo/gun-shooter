//go:build js

package ui

import (
	"log"
	"net/url"
	"strings"
	"syscall/js"

	"github.com/google/uuid"
)

var (
	document    = js.Global().Get("document")
	window      = js.Global().Get("window")
	location    = js.Global().Get("location")
	initialized = false
	params      url.Values
)

func init() {
	u, _ := url.Parse(location.Get("href").String())
	params = u.Query()
	if params.Get("id") == "" {
		uid, _ := uuid.NewV6()
		SetParam("id", uid.String())
	}
}

func Fullscreen(on bool) {
	log.Println("Fullscreen:", on)
	js.Global().Call("eval", `
			(function() {
				console.log('AudioContext resuming...');
				if (window.audioContext && window.audioContext.state === 'suspended') {
					window.audioContext.resume().then(() => {
						console.log('AudioContext resumed successfully');
					});
				}
				// 汎用的な検索
				document.querySelectorAll('audio, video').forEach(el => el.play().catch(() => {}));
			})()
		`)
	elm := document.Get("documentElement")
	if on {
		var f js.Func
		f = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			defer f.Release()
			window.Call("alert", "fullscreen error:", args[0])
			return nil
		})
		elm.Call("requestFullscreen").Call("catch", f)
	} else {
		if document.Get("fullscreenElement").Truthy() {
			go document.Call("exitFullscreen")
		}
	}
}

func BaseURL() string {
	return location.Get("origin").String() + location.Get("pathname").String()
}

func URLOpen(u string) {
	window.Call("open", u)
}

func GetParam(key string) string {
	return params.Get(key)
}

func SetParam(key, value string) {
	params.Set(key, value)
	location.Set("search", params.Encode())
}

func getViewFromHash() ViewName {
	return ViewName(strings.TrimPrefix(location.Get("hash").String(), "#"))
}

func initRouter(c *homeScreenComponent) {
	view := getViewFromHash()
	// Initial check
	if view != "" {
		log.Println("initial view", view)
		switch view {
		case ViewNamePlay:
			c.onPlayClicked()
		case ViewNameLicenses:
			c.onLicensesClicked()
		default:
			c.app.SetActiveView(view)
		}
	}
	initialized = true

	// Listen for changes
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		view := getViewFromHash()
		if view != c.app.ActiveView() {
			log.Println("view changed", view)
			switch view {
			case ViewNamePlay:
				c.onPlayClicked()
			case ViewNameLicenses:
				c.onLicensesClicked()
			default:
				c.app.SetActiveView(view)
			}
		}
		return nil
	})
	window.Call("addEventListener", "hashchange", cb)
}

func updateHash(view ViewName) {
	if !initialized {
		return
	}
	switch view {
	default:
		return
	case ViewNameHome, ViewNamePlay, ViewNameLicenses, ViewNameRoom:
	}
	targetHash := "#" + string(view)
	if location.Get("hash").String() != targetHash {
		log.Println("update hash", view)
		location.Set("hash", targetHash)
	}
}
