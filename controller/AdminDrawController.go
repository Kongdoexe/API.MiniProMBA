package controller

import (
	"fmt"

	"github.com/Kongdoexe/goland/database"
	"github.com/Kongdoexe/goland/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GenerateUniqueDraw(c *fiber.Ctx) error {
	var body struct {
		Rank   int  `json:"rank"`
		Status bool `json:"status"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "Cannot parse JSON"})
	}

	var selectPeriod models.LottoTicket
	if err := database.DBconn.Order("Period DESC").First(&selectPeriod).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถเรียกข้อมูลช่วงเวลาล่าสุดได้"})
	}

	// ตรวจสอบว่า rank ซ้ำหรือไม่
	var existingWinner models.Winner
	err := database.DBconn.Table("Winner").
		Joins("JOIN LottoTicket ON Winner.TicketID = LottoTicket.TicketID").
		Where("LottoTicket.Period = ? AND Winner.Rank = ?", selectPeriod.Period, body.Rank).
		First(&existingWinner).Error
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"msg": "มีอันดับนี้อยู่แล้วสำหรับช่วงเวลานี้"})
	} else if err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถตรวจสอบอันดับที่มีอยู่ได้"})
	}

	var selectedLotto models.LottoTicket
	query := database.DBconn.Table("LottoTicket").
		Where("Period = ?", selectPeriod.Period).
		Where("NOT EXISTS (SELECT 1 FROM Winner WHERE Winner.TicketID = LottoTicket.TicketID)")

	if !body.Status {
		query = query.Where("MemberID IS NOT NULL")
	}

	err = query.Order("RAND()").First(&selectedLotto).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบตั๋วล็อตโต้ที่มีคนซื้อ"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถเรียกข้อมูลตั๋วล็อตโต้"})
	}

	winner := models.Winner{
		Rank:     body.Rank,
		TicketID: selectedLotto.TicketID,
	}
	if err := database.DBconn.Create(&winner).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถสร้างผู้ชนะได้"})
	}

	dataout := struct {
		Rank     int    `json:"rank"`
		TicketID int    `json:"ticket_id"`
		Number   string `json:"number"`
	}{
		Rank:     body.Rank,
		TicketID: selectedLotto.TicketID,
		Number:   selectedLotto.Number,
	}

	return c.JSON(dataout)
}

func Getprize(c *fiber.Ctx) error {
	type Result struct {
		WinnerID uint
		Rank     uint
		MemberID *uint
		Gratuity float64
	}
	var selectPeriod models.LottoTicket
	if err := database.DBconn.Order("Period DESC").First(&selectPeriod).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถเรียกข้อมูลช่วงเวลาล่าสุดได้"})
	}

	var results []Result
	if err := database.DBconn.Table("Winner").
		Select("Winner.WinnerID, Winner.Rank, LottoTicket.MemberID, LottoTicket.Number, RankLotto.Gratuity").
		Joins("JOIN LottoTicket ON Winner.TicketID = LottoTicket.TicketID").
		Joins("JOIN RankLotto ON Winner.Rank = RankLotto.RankID").
		Where("LottoTicket.Period = ?", selectPeriod.Period).
		Order("Winner.WinnerID ASC").
		Scan(&results).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถเรียกข้อมูลผู้ชนะได้"})
	}

	return c.JSON(results)
}

func ResetSystem(c *fiber.Ctx) error {
	tables := []string{"Member", "LottoTicket", "Cart", "Winner"}

	// ลบข้อมูลในแต่ละตาราง
	for _, table := range tables {
		if err := database.DBconn.Exec(fmt.Sprintf("DELETE FROM `%s`", table)).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": fmt.Sprintf("Delete %s Failed.", table),
			})
		}

		if err := ResetAutoIncrement(database.DBconn, table); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": fmt.Sprintf("Reset AUTO_INCREMENT %s Failed.", table),
			})
		}
	}

	return c.JSON(fiber.Map{"msg": "Reset Success."})
}

func ResetAutoIncrement(db *gorm.DB, tableName string) error {
	// รันคำสั่ง SQL เพื่อรีเซ็ต AUTO_INCREMENT
	sql := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = 1", tableName)
	return db.Exec(sql).Error
}
