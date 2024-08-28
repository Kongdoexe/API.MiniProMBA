package controller

import (
	"github.com/Kongdoexe/goland/database"
	"github.com/Kongdoexe/goland/models"
	"github.com/gofiber/fiber/v2"
)

func AddFunds(c *fiber.Ctx) error {
	mid := c.Params("mid")
	var member models.Member

	var data struct {
		Gratuity int `json:"gratuity"`
	}

	if err := database.DBconn.First(&member, mid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบสมาชิก"})
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": err.Error()})
	}

	member.WalletBalance += data.Gratuity
	if err := database.DBconn.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถเพิ่มเงินได้"})
	}

	return c.JSON(fiber.Map{"msg": "เพิ่มเงินสำเร็จ"})
}
