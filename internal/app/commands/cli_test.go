package commands

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLevel(t *testing.T) {
	tmp := logrus.GetLevel()
	defer func() { logrus.SetLevel(tmp) }()

	err := logLevel("info").AfterApply()
	require.NoError(t, err)
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())
	err = logLevel("blah").AfterApply()
	require.Error(t, err)
}
