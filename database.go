package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

type ImdbDb struct {
	db *sql.DB
}

type ImdbTitleEntry struct {
	Id             string    `json:"id"`
	LastUpdate     string    `json:"lastUpdate"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	OriginalTitle  string    `json:"originalTitle"`
	Genres         string    `json:"genres"`
	Year           int       `json:"year"`
	ReleaseDate    string    `json:"releaseDate"`
	RuntimeMinutes int       `json:"runtimeMinutes"`
	IsAdult        bool      `json:"isAdult"`
	Rating         float64   `json:"rating"`
	Description    string    `json:"description"`
	ImageUrl       string    `json:"imageUrl"`
}

type ImdbTitleSearchFilter struct {
	Title     string
	MaxResult int
	Page      int
	SortBy    string
	Ascending bool
}

const DefaultDatabasePath = "./imdb.db"

func OpenDatabase(path string) (ImdbDb, error) {
	i:= ImdbDb{}
	db, err:= sql.Open("sqlite3", path)
	if err != nil {
		return i, err
	}
	i.db = db
	return i, nil
}

func OpenDefaultDatabase() (ImdbDb, error) {
	return OpenDatabase(DefaultDatabasePath)
}

func CreateDatabase(path string, deleteDb bool) (ImdbDb, error) {
	if deleteDb {
		if isFileExists(DefaultDatabasePath) {
			if err:= os.Remove(DefaultDatabasePath); err != nil {
				return ImdbDb{}, err
			}
		}
	}
	i, err:= OpenDatabase(path)
	if err != nil {
		return i, err
	}
	// Create initial table
	createSql := `
    DROP TABLE IF EXISTS titles;
	CREATE TABLE titles (
	    id TEXT NOT NULL PRIMARY KEY,
	    lastUpdate TEXT NOT NULL,
	    type TEXT NOT NULL,
	    title TEXT NOT NULL,
	    originalTitle TEXT,
	    genres TEXT,
	    year INTEGER,
	    releaseDate TEXT,
	    runtimeMinutes INTEGER,
	    isAdult INTEGER,
	    rating REAL,
	    description TEXT,
	    imageUrl TEXT
	);
	DELETE FROM titles;
	`
	if _, err = i.db.Exec(createSql); err != nil {
		i.Close()
		return i, err
	}
	return i, nil
}

func CreateDefaultDatabase(deleteDb bool) (ImdbDb, error) {
	return CreateDatabase(DefaultDatabasePath, deleteDb)
}

func IsDefaultDatabaseAvailable() bool {
	return isFileExists(DefaultDatabasePath)
}

func (i *ImdbDb) Close() error {
	return i.db.Close()
}

func (i *ImdbDb) Db() *sql.DB {
	return i.db
}

func (i *ImdbDb) PrepareInsertTitle() (*sql.Stmt, error) {
	insertSql:= `
    INSERT INTO titles (
        id,
        lastUpdate,
	    type,
	    title,
	    originalTitle,
	    genres,
	    year,
	    releaseDate,
	    runtimeMinutes,
	    isAdult,
	    rating,
	    description,
	    imageUrl) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?);
    `
	return i.db.Prepare(insertSql)
}

func (i *ImdbDb) InsertTitle(stmt *sql.Stmt, entry ImdbTitleEntry) error {
	_, err := stmt.Exec(
		entry.Id,
		entry.LastUpdate,
		entry.Type,
		entry.Title,
		entry.OriginalTitle,
		entry.Genres,
		entry.Year,
		entry.ReleaseDate,
		entry.RuntimeMinutes,
		entry.IsAdult,
		entry.Rating,
		entry.Description,
		entry.ImageUrl,
	)
	if err != nil {
		return err
	}
	return nil
}

func (i *ImdbDb) InsertTitleOnce(entry ImdbTitleEntry) error {
	stmt, err:= i.PrepareInsertTitle()
	if err != nil {
		return err
	}
	defer stmt.Close()
	err = i.InsertTitle(stmt, entry)
	if err != nil {
		return err
	}
	return nil
}

func (i *ImdbDb) GetTitlesCount(title string) int {
	count:= 0
	queryStr:= `SELECT COUNT(id) FROM titles WHERE title LIKE ?;`
	stmt, err:= i.db.Prepare(queryStr)
	if err != nil {
		return count
	}
	defer stmt.Close()
	title += "%"
	err = stmt.QueryRow(title).Scan(&count)
	if err != nil {
		count = 0
	}
	return count
}

func (i *ImdbDb) GetTitles(filter ImdbTitleSearchFilter) ([]ImdbTitleEntry, error) {
	var entries []ImdbTitleEntry = nil
	queryStr:= `SELECT * FROM titles WHERE title LIKE ? ORDER BY title ASC LIMIT ? OFFSET ?;`
	stmt, err:= i.db.Prepare(queryStr)
	if err != nil {
		return entries, err
	}
	defer stmt.Close()
	// default max result to 20 entries
	if filter.MaxResult == 0 {
		filter.MaxResult = 20
	}
	if filter.SortBy == "" {
		filter.SortBy = "title"
	}
	offset:= (filter.Page-1) * filter.MaxResult
	title:= filter.Title + "%"
	rows, err:= stmt.Query(title, filter.MaxResult, offset)
	if err != nil {
		return entries, err
	}
	for rows.Next() {
		entry:= ImdbTitleEntry{}
		err:= rows.Scan(
			&entry.Id,
			&entry.LastUpdate,
			&entry.Type,
			&entry.Title,
			&entry.OriginalTitle,
			&entry.Genres,
			&entry.Year,
			&entry.ReleaseDate,
			&entry.RuntimeMinutes,
			&entry.IsAdult,
			&entry.Rating,
			&entry.Description,
			&entry.ImageUrl,
			)
		if err != nil {
			return entries, err
		}
		entries = append(entries, entry)
	}
	if err:= rows.Err(); err != nil {
		return entries, err
	}
	return entries, nil
}

func (i *ImdbDb) GetTitleById(id string) (ImdbTitleEntry, error) {
	entry:= ImdbTitleEntry{}
	queryStr:= `SELECT * FROM titles WHERE id=?;`
	stmt, err:= i.db.Prepare(queryStr)
	if err != nil {
		return entry, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(id).Scan(
		&entry.Id,
		&entry.LastUpdate,
		&entry.Type,
		&entry.Title,
		&entry.OriginalTitle,
		&entry.Genres,
		&entry.Year,
		&entry.ReleaseDate,
		&entry.RuntimeMinutes,
		&entry.IsAdult,
		&entry.Rating,
		&entry.Description,
		&entry.ImageUrl,
		)
	return entry, err
}

func (i *ImdbDb) UpdateTitle(entry ImdbTitleEntry) error {
	queryStr:= `
    UPDATE titles 
    SET
        lastUpdate=?,
        type=?,
        title=?,
        originalTitle=?,
        genres=?,
        year=?,
        releaseDate=?,
        runtimeMinutes=?,
        isAdult=?,
        rating=?,
        description=?,
        imageUrl=?
    WHERE id=?;`
	stmt, err:= i.db.Prepare(queryStr)
	if err != nil {
		return err
	}
	defer stmt.Close()
	entry.LastUpdate = time.Now().String()
	_, err = stmt.Exec(
		entry.LastUpdate,
		entry.Type,
		entry.Title,
		entry.OriginalTitle,
		entry.Genres,
		entry.Year,
		entry.ReleaseDate,
		entry.RuntimeMinutes,
		entry.IsAdult,
		entry.Rating,
		entry.Description,
		entry.ImageUrl,
		entry.Id,
	)
	return err
}

func (i *ImdbDb) DeleteEntry(id string) error {
	queryStr:= `DELETE FROM titles WHERE id=?`
	stmt, err:= i.db.Prepare(queryStr)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}
