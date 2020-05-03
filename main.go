package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var DSN = "root:1234@tcp(localhost:3306)/dbgormendb?charset=utf8"

// dbe:{"table": "users"}
type User struct {
	ID       int    `dbe:"id,primary_key"`
	Login    string `dbe:"login,not_null"`
	Email    string
	Level    uint8
	IsActive bool
	UError   error `dbe:"-"`
}

func main() {
	var err error
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		fmt.Println("Unable to connect to DB", err)
		return
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("Unable to ping BD")
		return
	}
	newUser := &User{
		Login:    "newUser",
		Email:    "new@test.com",
		Level:    0,
		IsActive: false,
		UError:   nil,
	}

	err = newUser.createTable(db)
	if err != nil {
		fmt.Println("Error creating table.", err)

	}
	err = newUser.Create(db)
	if err != nil {
		fmt.Println("Error creating user.", err)
		return
	}

	nU := &User{}
	dbUsers, err := nU.Query(db)
	if err != nil {
		fmt.Println("Error selecting users.", err)
		return
	}
	fmt.Printf("From table users selected %d fields", len(dbUsers))
	var DBUser *User
	for _, user := range dbUsers {
		fmt.Println(user)
		DBUser = user
	}
	DBUser.Level = 2
	err = DBUser.Update(db)
	if err != nil {
		fmt.Println("Error updating users.", err)
		return
	}
	err = DBUser.Delete(db)
	if err != nil {
		fmt.Println("Error deleting users.", err)
		return
	}
}
