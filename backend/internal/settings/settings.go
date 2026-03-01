package settings

import (
	"database/sql"
	"strconv"
)

// Get retrieves a setting value by key. Returns empty string if not found.
func Get(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// GetInt retrieves a setting as int. Returns 0 if not found.
func GetInt(db *sql.DB, key string) (int, error) {
	value, err := Get(db, key)
	if err != nil {
		return 0, err
	}
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

// GetBool retrieves a setting as bool ("true"/"false"). Returns false if not found.
func GetBool(db *sql.DB, key string) (bool, error) {
	value, err := Get(db, key)
	if err != nil {
		return false, err
	}
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

// Set updates or inserts a setting.
func Set(db *sql.DB, key, value string) error {
	_, err := db.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = datetime('now')",
		key,
		value,
	)
	return err
}

// GetAll retrieves all settings as a map.
func GetAll(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return settings, nil
}
