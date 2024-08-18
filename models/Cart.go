package models

type Cart struct {
	CartID        int `gorm:"primaryKey;column:CartID;autoIncrement" json:"cartID"`
	MemberID      int `gorm:"column:MemberID" json:"memberID"`
	LottoTicketID int `gorm:"column:LottoTicketID" json:"lottoTicketID"`
}

func (Cart) TableName() string {
	return "Cart"
}
