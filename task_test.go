package htw_test

import (
	"testing"
	"time"

	"github.com/clockwerk-labs/htw"
	"github.com/stretchr/testify/require"
)

func TestTask_NewWithInt(t *testing.T) {
	inputTime := time.Now().Add(10 * time.Minute)
	inputValue := 42

	task := htw.NewTask(inputTime, inputValue)

	require.Equal(t, inputValue, task.Value)
	require.Equal(t, inputTime.UnixNano(), task.Expiration)
}
