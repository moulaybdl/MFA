package data

import (
	"database/sql"
	"errors"
)

// define the errors here:
var (
	ErrRecordNotFound = errors.New("record not found")
)

// define the Models struct 

type Models struct {
	Users UsersModel
}

func NewModel(db *sql.DB) Models {
	return Models{
		Users: UsersModel{DB: db},
	}

}