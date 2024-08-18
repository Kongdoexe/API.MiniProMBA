package models

type RankLotto struct {
	RankID   int     `gorm:"primaryKey;column:RankID;autoIncrement" json:"rankID"`
	Gratuity float64 `gorm:"column:Gratuity;type:int" json:"gratuity"`
}

func (RankLotto) TableName() string {
	return "RankLotto"
}
