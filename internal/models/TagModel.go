package models

type Tag struct { //他會自動改成snake_case然後給你加s 所以對照到了我sql的table name
	TagID     int       `gorm:"primaryKey" json:"tag_id"`
	TagName    string    `gorm:"primaryKey" json:"tag_name"`
}
