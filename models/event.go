package models

import "time"

// Example: Event = Name:XXXXX,Dept=OSS,EmplD:1234, Time=21-7-2021 21:00:10
type Event struct {
	Id    string       `json:"id"`
	Name  string       `json:"name"`
	Dept  string       `json:"dept"`
	EmpId int          `json:"empid"`
	Time   time.Time   `json:"time"`
}