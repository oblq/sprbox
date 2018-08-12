package sprbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepository(t *testing.T) {
	repo := NewRepository("./")
	repo.UpdateInfo()
	repo.PrintInfo()
	assert.NoError(t, repo.Error)
}

func TestNewWrongRepository(t *testing.T) {
	assert.Error(t, NewRepository("../").Error)
}
