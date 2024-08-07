package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thomas-introini/pocket-cli/models"
)

const DB_PATH = ".cache/pocket-cli-go/cache.db"

var NoUserErr = errors.New("user: no logged user found")
var NoSavesErr = errors.New("user: no saves found")

var DB *sql.DB

func ConnectDB() error {
	_, err := os.Stat(os.Getenv("HOME") + "/" + DB_PATH)
	if os.IsNotExist(err) {
		err = os.MkdirAll(os.Getenv("HOME")+"/"+filepath.Dir(DB_PATH), os.ModePerm)
		if err != nil {
			return err
		}
	}
	db, err := sql.Open("sqlite3", os.Getenv("HOME")+"/"+DB_PATH)
	if err != nil {
		return err
	}
	_, err = os.Stat(os.Getenv("HOME") + "/" + DB_PATH)
	if os.IsNotExist(err) {
		_, err = db.Exec(`
			CREATE TABLE user (
				username         TEXT PRIMARY KEY,
				access_token     TEXT,
				saves_updated_on INTEGER(8)
			)`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`
			CREATE TABLE save (
				id           TEXT PRIMARY KEY,
				title        TEXT,
				url          TEXT NOT NULL,
				description  TEXT,
				status       INTEGER(1),
				favorite     INTEGER(1),
			    tags         TEXT,
				time_to_read INTEGER,
				added_on     INTEGER(8),
				updated_on   INTEGER(8)
			)`)
		if err != nil {
			return err
		}
	}
	DB = db
	return nil
}

func GetLoggedUser() (user models.PocketUser, err error) {
	if DB == nil || DB.Ping() != nil {
		err = errors.New("could not connect to db")
		return
	}
	row := DB.QueryRow("SELECT access_token, username, saves_updated_on FROM user limit 1")
	var (
		accessToken string
		username    string
		updatedOn   int32 = 0
		i           sql.NullInt32
	)
	err = row.Scan(&accessToken, &username, &i)
	if err != nil {
		if err == sql.ErrNoRows { // authentication needed
			err = nil
			return
		} else {
			return
		}
	}

	if i.Valid {
		i.Scan(&updatedOn)
	}

	user = models.PocketUser{
		AccessToken:    accessToken,
		Username:       username,
		SavesUpdatedOn: updatedOn,
	}

	return
}

func GetPocketSaves() (list []models.PocketSave, err error) {
	list = make([]models.PocketSave, 0)
	rows, err := DB.Query(`
		SELECT id, title, url, description, time_to_read, status, favorite, tags, added_on, updated_on
		  FROM save
		 WHERE status = 0
		 ORDER BY added_on DESC`,
	)
	if err == sql.ErrNoRows {
		err = NoSavesErr
		return
	} else if err != nil {
		return
	}

	for rows.Next() {
		var (
			id         string
			title      string
			url        string
			desc       string
			timeToRead uint16
			status     uint8
			favorite   uint8
			tags       string
			addedOn    uint32
			updatedOn  uint32
		)
		if err = rows.Scan(
			&id,
			&title,
			&url,
			&desc,
			&timeToRead,
			&status,
			&favorite,
			&tags,
			&addedOn,
			&updatedOn,
		); err != nil {
			return
		}
		save := models.PocketSave{
			Id:              id,
			SaveTitle:       title,
			Url:             url,
			SaveDescription: desc,
			TimeToRead:      timeToRead,
			Status:          status,
			Favorite:        favorite == 1,
			Tags:            tags,
			AddedOn:         addedOn,
			UpdatedOn:       updatedOn,
		}
		list = append(list, save)
	}
	return
}

func SaveUser(accessToken, username string) (models.PocketUser, error) {
	current, err := GetLoggedUser()
	if err != nil {
		return models.NoUser, err
	}
	if current == models.NoUser {
		_, err = DB.Exec("INSERT INTO user(access_token, username) VALUES (?,?)", accessToken, username)
	} else {
		_, err = DB.Exec("UPDATE user SET access_token = ? WHERE user = ?", accessToken, username)
	}
	if err != nil {
		fmt.Println("could not save user", err)
		return models.NoUser, err
	}
	user := models.PocketUser{
		AccessToken:    accessToken,
		Username:       username,
		SavesUpdatedOn: 0,
	}
	fmt.Println("saved", user)
	return user, err
}

func InsertSaves(since float64, saves []models.PocketSave) ([]models.PocketSave, error) {
	ret := make([]models.PocketSave, 0)
	tx, err := DB.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: false})
	if err != nil {
		defer tx.Rollback()
		return ret, err
	}
	for _, save := range saves {
		if save.Status == models.StatusDeleted {
			_, err = tx.Exec(
				"DELETE FROM save where id = ?",
				save.Id,
			)
		} else {
			_, err = tx.Exec(
				`INSERT INTO save(id, title, url, description, time_to_read, status, favorite, tags, added_on, updated_on)
			 VALUES(?,?,?,?,?,?,?,?,?,?)
			 ON CONFLICT(id) DO
			 UPDATE SET
			  title = excluded.title,
				url = excluded.url,
		description = excluded.description,
	   time_to_read = excluded.time_to_read,
			 status = excluded.status,
		   favorite = excluded.favorite,
			   tags = excluded.tags,
		   added_on = excluded.added_on,
		 updated_on = excluded.updated_on`,
				save.Id,
				save.SaveTitle,
				save.Url,
				save.SaveDescription,
				save.TimeToRead,
				save.Status,
				save.Favorite,
				save.Tags,
				save.AddedOn,
				save.UpdatedOn,
			)
		}
		if err != nil {
			defer tx.Rollback()
			return ret, err
		}
		ret = append(ret, save)
	}
	_, err = tx.Exec("UPDATE user SET saves_updated_on = ?", since)
	if err != nil {
		defer tx.Rollback()
		return ret, err
	}
	err = tx.Commit()
	if err != nil {
		defer tx.Rollback()
		return ret, err
	}
	return ret, nil
}
