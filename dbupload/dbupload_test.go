package dbupload

import "testing"

func TestTime(t *testing.T) {
	if err := Upload(nil, nil); err != nil {
		t.Error(err)
	}
}
