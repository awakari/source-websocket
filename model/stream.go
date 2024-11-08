package model

import "time"

type Stream struct {
	CreatedAt time.Time
	Request   string
	GroupId   string
	UserId    string
	Replica   uint32
}
