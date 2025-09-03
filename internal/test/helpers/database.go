package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type TestDatabase struct {
	DB  *sql.DB
	DSN string
}

func SetupTestDatabase(t *testing.T) *TestDatabase {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/resource_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Failed to connect to test database")

	err = db.Ping()
	require.NoError(t, err, "Failed to ping test database")

	return &TestDatabase{
		DB:  db,
		DSN: dsn,
	}
}

func (td *TestDatabase) Cleanup(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	// Clean up test data
	tables := []string{
		"achievements",
		"resource_uploads",
	}

	for _, table := range tables {
		_, err := td.DB.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}

func (td *TestDatabase) Close() error {
	return td.DB.Close()
}

func (td *TestDatabase) SeedAchievements(t *testing.T, achievements []map[string]interface{}) {
	t.Helper()

	ctx := context.Background()

	for _, achievement := range achievements {
		query := `
			INSERT INTO achievements (id, name, description, category, points, is_active, createdAt, updatedAt)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		`

		_, err := td.DB.ExecContext(ctx, query,
			achievement["id"],
			achievement["name"],
			achievement["description"],
			achievement["category"],
			achievement["points"],
			achievement["is_active"],
		)
		require.NoError(t, err, "Failed to seed achievement")
	}
}

func (td *TestDatabase) CountAchievements(t *testing.T) int {
	t.Helper()

	var count int
	err := td.DB.QueryRow("SELECT COUNT(*) FROM achievements").Scan(&count)
	require.NoError(t, err, "Failed to count achievements")

	return count
}

func (td *TestDatabase) GetAchievementByID(t *testing.T, id string) map[string]interface{} {
	t.Helper()

	row := td.DB.QueryRow(`
		SELECT id, name, description, category, points, is_active
		FROM achievements
		WHERE id = $1
	`, id)

	var achievement = make(map[string]interface{})
	var achievementID, name, description, category string
	var points int
	var isActive bool

	err := row.Scan(&achievementID, &name, &description, &category, &points, &isActive)
	require.NoError(t, err, "Failed to get achievement")

	achievement["id"] = achievementID
	achievement["name"] = name
	achievement["description"] = description
	achievement["category"] = category
	achievement["points"] = points
	achievement["is_active"] = isActive

	return achievement
}
