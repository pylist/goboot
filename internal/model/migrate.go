package model

import "goboot/pkg/database"

func AutoMigrate() error {
	return database.DB.AutoMigrate(
		&User{},
	)
}
