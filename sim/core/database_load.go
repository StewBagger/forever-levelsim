//go:build with_db

// Only include this file in the build when we specify the 'with_db' tag.
// Without the tag, the database will start out completely empty.
package core

import (
	"github.com/wowsims/forever/assets/database"
)

func init() {
	WITH_DB = true
	addToDatabase(SimDatabaseFromUI(database.Load()))
}
