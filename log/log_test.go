package log_test

import (
	"testing"

	"github.com/icexin/eggos/log"
)

func ExampleSetLogLevel() {
	// Set the log level to Debug
	log.Level = log.LoglvlDebug
}

func TestLog_SetLevel(t *testing.T) {
	for _, test := range []struct {
		name        string
		l           log.LogLevel
		expectError bool
	}{
		// These two cases require users to do something pretty explcitly
		// wrong to hit, but they're worth catching
		{"log level is too low", log.LogLevel(-1), true},
		{"log level is too high", log.LogLevel(6), true},

		// Included log levels
		{"log.LoglvlDebug is valid", log.LoglvlDebug, false},
		{"log.LoglvlInfo is valid", log.LoglvlInfo, false},
		{"log.LoglvlWarn is valid", log.LoglvlWarn, false},
		{"log.LoglvlError is valid", log.LoglvlError, false},
		{"log.LoglvlNone is valid", log.LoglvlNone, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := log.SetLevel(test.l)

			if err == nil && test.expectError {
				t.Error("expected error, received none")
			} else if err != nil && !test.expectError {
				t.Errorf("unexpected error: %#v", err)
			}

			if !test.expectError {
				if log.Level != test.l {
					t.Errorf("log.Level should be %#v, received %#v", log.Level, test.l)
				}
			}
		})
	}
}
