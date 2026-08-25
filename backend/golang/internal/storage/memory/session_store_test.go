package memory_test

import (
	"errors"
	"testing"

	"github.com/romanthekat/planning-poker/internal/poker"
	"github.com/romanthekat/planning-poker/internal/storage/memory"
)

func TestStore_CreateAndGet(t *testing.T) {
	store := memory.NewStore()

	created, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := store.Get(created.Id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Id != created.Id {
		t.Errorf("Get() id = %v, want %v", got.Id, created.Id)
	}
	if got != created {
		t.Errorf("Get() returned a different *Session instance than Create()")
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	store := memory.NewStore()

	_, err := store.Get(poker.SessionId(123456))
	if !errors.Is(err, poker.ErrNoRecord) {
		t.Errorf("Get() error = %v, want %v", err, poker.ErrNoRecord)
	}
}

func TestStore_Delete(t *testing.T) {
	store := memory.NewStore()

	session, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := store.Delete(session.Id); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = store.Get(session.Id)
	if !errors.Is(err, poker.ErrNoRecord) {
		t.Errorf("Get() after Delete() error = %v, want %v", err, poker.ErrNoRecord)
	}
}

func TestStore_Delete_NonExistentIsNoop(t *testing.T) {
	store := memory.NewStore()

	if err := store.Delete(poker.SessionId(999999)); err != nil {
		t.Errorf("Delete() of non-existent session error = %v, want nil", err)
	}
}

func TestStore_List(t *testing.T) {
	store := memory.NewStore()

	if list := store.List(); len(list) != 0 {
		t.Fatalf("List() on empty store = %v, want empty", list)
	}

	first, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	second, err := store.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	list := store.List()
	if len(list) != 2 {
		t.Fatalf("List() len = %d, want 2", len(list))
	}

	seen := map[poker.SessionId]bool{}
	for _, s := range list {
		seen[s.Id] = true
	}
	if !seen[first.Id] || !seen[second.Id] {
		t.Errorf("List() = %v, want to contain %v and %v", list, first.Id, second.Id)
	}
}
