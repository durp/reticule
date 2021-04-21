package coinbasepro

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	err := Error{
		StatusCode: 403,
		Message:    "verboten",
	}
	assert.Equal(t, "verboten (403)", err.Error())
}

func TestCapture(t *testing.T) {
	deferredErr := errors.New("catch me if you can")
	t.Run("CapturedErrorEnsuresDeferredErrorsCaught", func(t *testing.T) {
		captured := func() (capture error) {
			defer func() {
				Capture(&capture, func() error {
					return deferredErr
				}())
			}()
			return nil
		}()
		assert.Equal(t, deferredErr, captured)
	})
	t.Run("CapturedErrorDoesNotStompOnExistingErrors", func(t *testing.T) {
		alreadyFailed := errors.New("already failed")
		captured := func() (capture error) {
			defer func() {
				Capture(&capture, func() error {
					return deferredErr
				}())
			}()
			return alreadyFailed
		}()
		assert.Equal(t, alreadyFailed, captured)
	})
}
