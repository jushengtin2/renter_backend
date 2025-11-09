package models

type PostTag struct { //他會自動改成snake_case然後給你加s 所以對照到了我sql的table name
	TagID     int       `gorm:"primaryKey" json:"tag_id"`
	PostID    int    `gorm:"primaryKey" json:"post_id"`
}
