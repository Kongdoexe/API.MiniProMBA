package controller

import (
	"github.com/Kongdoexe/goland/database"
	"github.com/Kongdoexe/goland/models"
	"github.com/gofiber/fiber/v2"
)

func GetSalesData(c *fiber.Ctx) error {
	salesSummary, err := getSalesSummaryData()
	if err != nil {
		return c.Status(500).SendString("ไม่สามารถนับจำนวนทั้งหมดได้")
	}

	soldNumbersCount, err := getSoldNumbersCountData()
	if err != nil {
		return c.Status(500).SendString("ไม่สามารถนับจำนวนยอดขายได้")
	}

	remainingNumbersCount, err := getRemainingNumbersCountData()
	if err != nil {
		return c.Status(500).SendString("ไม่สามารถรับการนับจำนวนที่เหลือได้")
	}

	data := map[string]interface{}{
		"salesSummary":          salesSummary,
		"soldNumbersCount":      soldNumbersCount,
		"remainingNumbersCount": remainingNumbersCount,
	}

	return c.JSON(data)
}

func getSalesSummaryData() (int64, error) {
	var count int64
	var selectPeriod models.LottoTicket

	if err := database.DBconn.Order("Period DESC").First(&selectPeriod).Error; err != nil {
		return 0, err
	}
	err := database.DBconn.Model(&models.LottoTicket{}).Where("Period = ?", selectPeriod.Period).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func getSoldNumbersCountData() (int64, error) {
	var count int64
	err := database.DBconn.Model(&models.LottoTicket{}).Where("MemberID IS NOT NULL").Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func getRemainingNumbersCountData() (int64, error) {
	var count int64
	var selectPeriod models.LottoTicket

	if err := database.DBconn.Order("Period DESC").First(&selectPeriod).Error; err != nil {
		return 0, err
	}

	err := database.DBconn.Model(&models.LottoTicket{}).Where("MemberID IS NULL AND Period = ?", selectPeriod.Period).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}
