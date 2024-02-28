package migrations

import (
	"gitlab.com/clseibold/auragem_sis/db"
	"gitlab.com/clseibold/auragem_sis/migration/types"
)

var All map[types.MigrationVersion]types.Migration = make(map[types.MigrationVersion]types.Migration)
var Music map[types.MigrationVersion]types.Migration = make(map[types.MigrationVersion]types.Migration)
var LifeKept map[types.MigrationVersion]types.Migration = make(map[types.MigrationVersion]types.Migration)
var StarWars map[types.MigrationVersion]types.Migration = make(map[types.MigrationVersion]types.Migration)
var Search map[types.MigrationVersion]types.Migration = make(map[types.MigrationVersion]types.Migration)
var Ask map[types.MigrationVersion]types.Migration = make(map[types.MigrationVersion]types.Migration)

func GetMap(database db.DBType) map[types.MigrationVersion]types.Migration {
	if database == db.MusicDB {
		return Music
	} else if database == db.LifeKeptDB {
		return LifeKept
	} else if database == db.StarWarsDB {
		return StarWars
	} else if database == db.SearchDB {
		return Search
	} else if database == db.AskDB {
		return Ask
	}

	panic("Have you added the database to migrations.go?")
}

func registerMigration(m types.Migration) {
	All[m.Version()] = m

	database := m.DB()
	if database == db.MusicDB {
		Music[m.Version()] = m
	} else if database == db.LifeKeptDB {
		LifeKept[m.Version()] = m
	} else if database == db.StarWarsDB {
		StarWars[m.Version()] = m
	} else if database == db.SearchDB {
		Search[m.Version()] = m
	} else if database == db.AskDB {
		Ask[m.Version()] = m
	} else {
		panic("Have you added the database to migrations.go?")
	}
}
