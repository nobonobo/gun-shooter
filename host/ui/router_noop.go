//go:build !js

package ui

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

func Fullscreen(on bool) {}

func BaseURL() string {
	return "https://nobonobo.github.io/gun-shooter/"
}

func URLOpen(u string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Windows: rundll32またはstartを使用
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	case "darwin": // macOS
		cmd = exec.Command("open", u)
	case "linux":
		cmd = exec.Command("xdg-open", u)
	default:
		log.Println(fmt.Errorf("対応していないOS: %s", runtime.GOOS))
	}
	cmd.Start()
}

func GetParam(key string) string {
	return ""
}

func SetParam(key, value string) {
}

func getViewFromHash() ViewName {
	return ""
}

func initRouter(app *applicationComponent) {
}

func updateHash(view ViewName) {
}
