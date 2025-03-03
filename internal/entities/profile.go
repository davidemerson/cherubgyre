package entities

import (
	"time"

	"github.com/uptrace/bun"
)

type ProfileEntity struct {
	bun.BaseModel `bun:"table:profiles" json:"-"`
	Id            *string    `bun:"id,pk,autoincrement" json:""`
	PrivateEmail  *string    `bun:"private_email" json:"private_email"`
	ProfileImage  *string    `bun:"profile_image" json:"profile_image"`
	Fullname      *string    `bun:"fullname" json:"fullname"`
	ExternalId    *string    `bun:"external_id" json:"external_id"`
	Created       *time.Time `bun:"created,notnull,default:now()" json:"created"`
	Updated       *time.Time `bun:"updated,notnull,default:now()" json:"updated"`
}
