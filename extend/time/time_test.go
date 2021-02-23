package cherryTime

import "testing"

func TestCherryTime_Now(t *testing.T) {
	AddGlobalOffset(100)

	now := Now()
	t.Log(now.String())

}

func TestCherryTime_Yesterday(t *testing.T) {
	yesterday := Yesterday()
	t.Log(yesterday.String())
}
