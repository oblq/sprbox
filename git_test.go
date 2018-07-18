package sprbox

import "testing"

func TestNewRepository(t *testing.T) {
	repo := NewRepository("./")
	repo.UpdateInfo()
	repo.PrintInfo()
	if repo.Error != nil {
		t.Fail()
	}
}

func TestNewWrongRepository(t *testing.T) {
	repo := NewRepository("../")
	if repo.Error == nil {
		t.Fail()
	}
}
