package models

type LottoTicket struct {
	TicketID int    `gorm:"primaryKey;column:TicketID;autoIncrement" json:"TicketID"`
	MemberID *int   `gorm:"null;column:MemberID" json:"memberID"`
	Number   string `gorm:"column:Number;type:varchar(6)" json:"number"`
	Period   int    `gorm:"column:Period;type:int" json:"period"`
	Price    int    `gorm:"column:Price;type:int" json:"price"`
}

func (LottoTicket) TableName() string {
	return "LottoTicket"
}
