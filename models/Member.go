package models

type Member struct {
	MemberID      int    `gorm:"primaryKey;autoIncrement;column:MemberID" json:"MemberID"`
	Name          string `gorm:"type:varchar(50);column:Name" json:"name"`
	Email         string `gorm:"type:varchar(50);column:Email" json:"email"`
	Password      string `gorm:"type:text;column:Password" json:"password"`
	Phone         int    `gorm:"type:int;column:Phone" json:"phone"`
	WalletBalance int    `gorm:"type:int;column:WalletBalance" json:"wallet"`
	IsAdmin       int    `gorm:"type:int;column:IsAdmin" json:"isadmin"`
}

func (Member) TableName() string {
	return "Member"
}
