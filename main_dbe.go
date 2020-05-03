package main

import "database/sql"

func (in *User) createTable(db *sql.DB) error {
	sqlQ := `CREATE TABLE users (
id integer NOT NULL AUTO_INCREMENT,
login text NOT NULL,
Email text,
Level integer,
IsActive boolean,
PRIMARY KEY (id)
)`
	_, err := db.Exec(sqlQ)
	if err != nil {
		return err
	}
	return nil
}

func (in *User) Create(db *sql.DB) error {
	sqlQ := "INSERT INTO users (`login`,`Email`,`Level`,`IsActive`) VALUES (?,?,?,?);"

	result, err := db.Exec(sqlQ, in.Login, in.Email, in.Level, in.IsActive)
	if err != nil {
		return err
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		return nil
	}
	in.ID = int(lastId)
	return nil
}

func (in *User) Query(db *sql.DB) ([]*User, error) {
	sqlQ := "SELECT * FROM users;"
	rows, err := db.Query(sqlQ)
	results := make([]*User, 0)
	for rows.Next() {
		tempR := &User{}
		err = rows.Scan(&tempR.ID, &tempR.Login, &tempR.Email, &tempR.Level, &tempR.IsActive)
		if err != nil {
			return nil, err
		}
		results = append(results, tempR)
	}
	return results, nil
}

func (in *User) Update(db *sql.DB) error {
	sqlQ := "UPDATE users SET `login`=?,`Email`=?,`Level`=?,`IsActive`=? WHERE id = ?;"

	_, err := db.Exec(sqlQ, in.Login, in.Email, in.Level, in.IsActive, in.ID)
	if err != nil {
		return err
	}
	return nil
}

func (in *User) Delete(db *sql.DB) error {
	sqlQ := "DELETE FROM users WHERE id = ?"
	_, err := db.Exec(sqlQ, in.ID)
	if err != nil {
		return err
	}
	return nil
}
