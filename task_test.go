package htw_test

import (
	"testing"
	"time"

	"github.com/clockwerk-labs/htw"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTask_NewWithInt(t *testing.T) {
	inputID := uuid.New()
	inputTime := time.Now().Add(10 * time.Minute)
	inputValue := 42

	task := htw.NewTask(inputID, inputTime, inputValue)

	require.Equal(t, inputID, task.Id)
	require.Equal(t, inputValue, task.Value)
	require.Equal(t, inputTime.UnixNano(), task.Expiration)
}
