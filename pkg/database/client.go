package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"
)

const darkRoomDir = ".darkroom"
const dbFileName = "db.sqlite"

func InitClient() (*sqlx.DB, error) {
	appDir, err := getOrCreateAppDir()
	if err != nil {
		return nil, err
	}
	dbPath := path.Join(appDir, dbFileName)
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(schemaDefinition)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getOrCreateAppDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	appDir := path.Join(homeDir, darkRoomDir)
	_, err = os.ReadDir(appDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(appDir, 0755)
			if err != nil {
				return "", err
			}
			_, err = os.ReadDir(appDir)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return appDir, nil
}
