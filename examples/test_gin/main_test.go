package main

import (
	"fmt"
	cherryHttp "github.com/cherry-game/cherry/extend/http"
	"net/http"
	"testing"
	"time"
)

func TestControllerMaxConnect(t *testing.T) {

	for i := 0; i < 100; i++ {
		go func(i int) {
			result, rsp, _ := cherryHttp.GET("http://127.0.0.1:10820")

			if rsp != nil && rsp.StatusCode != http.StatusOK {
				fmt.Printf("index = %d, result = %s, code = %v\n", i, result, rsp.StatusCode)
			}
		}(i)
	}

	time.Sleep(1 * time.Hour)
}
