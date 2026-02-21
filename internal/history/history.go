package history

import "local-clipboard/internal/models"

// History provides persistence for clipboard entries.
type History interface {
	Init() error
	Insert(text, source string) (models.ClipboardUpdate, error)
	Latest() (models.ClipboardUpdate, error)
	ByID(id int64) (models.ClipboardUpdate, error)
	List(limit int, search string) ([]models.ClipboardUpdate, error)
	SetPinned(id int64, pinned bool) error
	Delete(id int64) error
}
