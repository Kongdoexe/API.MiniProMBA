package controller

import (
	"github.com/Kongdoexe/goland/database"
	"github.com/Kongdoexe/goland/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Register(c *fiber.Ctx) error {
	var member models.Member

	// ผูก Body ที่ส่งเข้ามาเข้ากับ member
	if err := c.BodyParser(&member); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": err.Error()})
	}

	// สร้าง query ส่งเข้าไปใน database และเก็บไว้ใน existingMember
	var existingMember models.Member
	database.DBconn.Where("email = ?", member.Email).First(&existingMember)
	if existingMember.MemberID != 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"msg": "มีอีเมลอยู่แล้ว"})
	}

	// ทำการ hash รหัสผ่าน
	hashedPassword, err := HashPassword(member.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถแฮชรหัสผ่านได้"})
	}
	// เปลี่ยนรหัสผ่าน
	member.Password = hashedPassword

	// สร้าง user ใหม่
	database.DBconn.Create(&member)
	return c.JSON(member)
}

func Login(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": err.Error()})
	}

	var memberFind models.Member
	database.DBconn.Where("email = ?", input.Email).First(&memberFind)

	if memberFind.MemberID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "อีเมลไม่ถูกต้อง"})
	}

	if !CheckPasswordHash(input.Password, memberFind.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "รหัสผ่านไม่ถูกต้อง"})
	}

	return c.JSON(fiber.Map{"MemberID": memberFind.MemberID})
}

func UpdateProfile(c *fiber.Ctx) error {
	mid := c.Params("mid")
	var member models.Member

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone int    `json:"phone"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "ไม่สามารถแยกวิเคราะห์เนื้อหาคำขอ"})
	}

	if err := database.DBconn.First(&member, mid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบสมาชิก"})
	}

	var existingMember models.Member
	if err := database.DBconn.Where("Email = ? AND MemberID != ?", input.Email, mid).First(&existingMember).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"msg": "อีเมลมีการใช้งานโดยสมาชิกรายอื่นแล้ว"})
	}

	if err := database.DBconn.Where("Phone = ? AND MemberID != ?", input.Phone, mid).First(&existingMember).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"msg": "โทรศัพท์มีการใช้งานโดยสมาชิกรายอื่นแล้ว"})
	}

	member.Name = input.Name
	member.Email = input.Email
	member.Phone = input.Phone

	if err := database.DBconn.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถอัปเดตโปรไฟล์ได้"})
	}

	return c.JSON(member)
}

func ChangePassword(c *fiber.Ctx) error {
	mid := c.Params("mid")
	var member models.Member
	var data struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := database.DBconn.First(&member, mid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบสมาชิก"})
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": err.Error()})
	}

	if data.OldPassword == data.NewPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "รหัสผ่านเก่าและใหม่เหมือนกัน"})
	}

	if !CheckPasswordHash(data.OldPassword, member.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "รหัสผ่านปัจจุบันไม่ถูกต้อง"})
	}

	hashedPassword, err := HashPassword(data.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ล้มเหลวในการแฮรหัสผ่าน"})
	}

	member.Password = hashedPassword
	if err := database.DBconn.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถอัปเดตรหัสผ่าน"})
	}

	return c.JSON(fiber.Map{"msg": "อัปเดตรหัสผ่านเรียบร้อยแล้ว"})
}

func DeleteAccount(c *fiber.Ctx) error {
	mid := c.Params("mid")
	var member models.Member

	result := database.DBconn.Delete(&member, "MemberID = ?", mid)

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "ไม่พบสมาชิก",
		})
	}

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "ลบสมาชิกไม่สำเร็จ",
		})
	}

	return c.JSON(fiber.Map{"msg": "ลบสมาชิกสำเร็จ"})
}
