package models

type Winner struct {
	WinnerID int `gorm:"primaryKey;column:WinnerID;autoIncrement" json:"winnerID"`
	TicketID int `gorm:"column:TicketID" json:"ticketID"`
	Rank     int `gorm:"column:Rank" json:"rank"`
}

func (Winner) TableName() string {
	return "Winner"
}
