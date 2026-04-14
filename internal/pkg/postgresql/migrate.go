package postgresql

import "fmt"

func Migrate(database *Database) error {
	err := database.db.AutoMigrate(
		new(ProjectModel),
		new(BacklogItemModel),
	)
	if err != nil {
		return fmt.Errorf("migrate postgres schema: %w", err)
	}

	return nil
}
