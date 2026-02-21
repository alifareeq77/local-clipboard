package history

import (
	"encoding/json"
	"fmt"
	"local-clipboard/internal/models"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SqliteHistory implements History using the sqlite3 CLI.
type SqliteHistory struct {
	path string
	mu   sync.Mutex
}

// NewSqlite returns a new SqliteHistory that uses the given database path.
func NewSqlite(path string) *SqliteHistory {
	return &SqliteHistory{path: path}
}

// Init creates the clipboard_history table if needed and runs the pinned column migration.
func (s *SqliteHistory) Init() error {
	_, err := s.runSQL(`CREATE TABLE IF NOT EXISTS clipboard_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		source TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		pinned INTEGER NOT NULL DEFAULT 0
	);
	ALTER TABLE clipboard_history ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0;`)
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "duplicate column name") {
		return nil
	}
	return err
}

// Insert adds a new clipboard entry and returns it with ID and timestamps.
// Query is built by concatenation (not fmt.Sprintf) so user text containing '%' cannot break the SQL.
func (s *SqliteHistory) Insert(text, source string) (models.ClipboardUpdate, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	query := "INSERT INTO clipboard_history(text,source,updated_at,pinned) VALUES(" +
		sqlQuoteMultiline(text) + "," +
		sqlQuoteMultiline(source) + "," +
		sqlQuote(now) + ",0); SELECT last_insert_rowid();"
	out, err := s.runSQL(query)
	if err != nil {
		return models.ClipboardUpdate{}, err
	}
	// sqlite3 may output a line per statement; last_insert_rowid() is on the last line.
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var lastLine string
	for i := len(lines) - 1; i >= 0; i-- {
		if s := strings.TrimSpace(lines[i]); s != "" {
			lastLine = s
			break
		}
	}
	if lastLine == "" {
		return models.ClipboardUpdate{}, errNotFound
	}
	id, err := strconv.ParseInt(lastLine, 10, 64)
	if err != nil || id <= 0 {
		if err == nil {
			err = errNotFound
		}
		return models.ClipboardUpdate{}, err
	}
	// Return the row we just inserted; don't use ByID — selectRows with .mode tabs
	// breaks when text contains newlines/tabs, so we'd get "not found".
	updatedAt, _ := time.Parse(time.RFC3339Nano, now)
	return models.ClipboardUpdate{
		ID:        id,
		Text:      text,
		Source:    source,
		UpdatedAt: updatedAt,
		Pinned:    false,
	}, nil
}

// SetPinned sets the pinned flag for the given entry.
func (s *SqliteHistory) SetPinned(id int64, pinned bool) error {
	pinInt := 0
	if pinned {
		pinInt = 1
	}
	_, err := s.runSQL(fmt.Sprintf("UPDATE clipboard_history SET pinned=%d WHERE id=%d;", pinInt, id))
	return err
}

// Delete removes the entry with the given id.
func (s *SqliteHistory) Delete(id int64) error {
	_, err := s.runSQL(fmt.Sprintf("DELETE FROM clipboard_history WHERE id=%d;", id))
	return err
}

// Latest returns the most recent entry (pinned first, then by id).
func (s *SqliteHistory) Latest() (models.ClipboardUpdate, error) {
	rows, err := s.selectRows("SELECT id,text,source,updated_at,pinned FROM clipboard_history ORDER BY pinned DESC, id DESC LIMIT 1;")
	if err != nil || len(rows) == 0 {
		if err == nil {
			err = errNoRows
		}
		return models.ClipboardUpdate{}, err
	}
	return rows[0], nil
}

// ByID returns the entry with the given id.
func (s *SqliteHistory) ByID(id int64) (models.ClipboardUpdate, error) {
	rows, err := s.selectRows(fmt.Sprintf("SELECT id,text,source,updated_at,pinned FROM clipboard_history WHERE id=%d LIMIT 1;", id))
	if err != nil || len(rows) == 0 {
		if err == nil {
			err = errNotFound
		}
		return models.ClipboardUpdate{}, err
	}
	return rows[0], nil
}

// List returns up to limit entries, optionally filtered by search (LIKE on text).
func (s *SqliteHistory) List(limit int, search string) ([]models.ClipboardUpdate, error) {
	query := "SELECT id,text,source,updated_at,pinned FROM clipboard_history"
	if search != "" {
		query += " WHERE lower(text) LIKE " + sqlQuote("%"+strings.ToLower(search)+"%")
	}
	query += fmt.Sprintf(" ORDER BY pinned DESC, id DESC LIMIT %d;", limit)
	return s.selectRows(query)
}

func (s *SqliteHistory) selectRows(query string) ([]models.ClipboardUpdate, error) {
	out, err := s.runSQL(".mode json\n" + query)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return []models.ClipboardUpdate{}, nil
	}
	var raw []struct {
		ID        int64  `json:"id"`
		Text      string `json:"text"`
		Source    string `json:"source"`
		UpdatedAt string `json:"updated_at"`
		Pinned    int    `json:"pinned"`
	}
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("sqlite json: %w", err)
	}
	rows := make([]models.ClipboardUpdate, 0, len(raw))
	for _, r := range raw {
		t, _ := time.Parse(time.RFC3339Nano, r.UpdatedAt)
		rows = append(rows, models.ClipboardUpdate{
			ID:        r.ID,
			Text:      r.Text,
			Source:    r.Source,
			UpdatedAt: t,
			Pinned:    r.Pinned != 0,
		})
	}
	return rows, nil
}

func (s *SqliteHistory) runSQL(sql string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cmd := exec.Command("sqlite3", s.path)
	cmd.Stdin = strings.NewReader(sql)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("sqlite3: %w: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func sqlQuote(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

func sqlQuoteMultiline(v string) string {
	v = strings.ReplaceAll(v, "\x00", "")
	v = strings.ReplaceAll(v, "\r\n", "\n")
	v = strings.ReplaceAll(v, "\r", "\n")
	if v == "" {
		return "''"
	}
	parts := strings.Split(v, "\n")
	for i := range parts {
		parts[i] = "'" + strings.ReplaceAll(parts[i], "'", "''") + "'"
	}
	return strings.Join(parts, "||char(10)||")
}

var errNoRows = fmt.Errorf("no rows")
var errNotFound = fmt.Errorf("not found")
