package model

type User struct {
	BaseModel
	Username string `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password string `gorm:"size:255;not null" json:"-"`
	Nickname string `gorm:"size:50" json:"nickname"`
	Phone    string `gorm:"size:20;index" json:"phone"`
	Email    string `gorm:"size:100;index" json:"email"`
	Avatar   string `gorm:"size:255" json:"avatar"`
	Status   int8   `gorm:"default:1" json:"status"` // 1: active, 0: disabled
	Role     int8   `gorm:"default:0" json:"role"`   // 0: user, 1: admin
}

func (User) TableName() string {
	return "users"
}
