package forecast

import (
	"testing"
)

func TestRequestForecast(t *testing.T) {
	res := Request(130010)

	if len(res) == 0 {
		t.Errorf("result none.")
	}
}
