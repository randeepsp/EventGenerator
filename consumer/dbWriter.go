package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/Shopify/sarama"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB
func IntiailizeDB(ctx context.Context) error{
	log.Println("initialise db")
	var err error
	db, err = sql.Open("mysql", "root:pass1@tcp(127.0.0.1:3306)/events")
	if err!= nil {
		log.Printf("failed to connect to db due to %v", err)
		return err
	}

	return nil
}

func pushMsgtoDB(msg *sarama.ConsumerMessage){
	//unmarshal
	event := Event{}
	err := json.Unmarshal(msg.Value,&event)
	if err != nil {
		log.Printf("error %s unmarshalling to event struct ", err)
	}
	//push to db
	err = writeToDB(event)
	if err != nil {
		log.Printf("error inserting %s to db ", err)
	}
}


func writeToDB(event Event) error{
	// perform a db.Query insert
	insertStmt, err := db.Prepare("INSERT INTO events(id,name,dept,empid,etime) VALUES(?,?,?,?,?)")
	if err!= nil {
		log.Printf("unable to insert event into db due to %v", err)
		return err
	}
	defer insertStmt.Close()
	insertStmt.Exec(event.Id, event.Name, event.Dept, event.EmpId, event.Time)
	return nil
}