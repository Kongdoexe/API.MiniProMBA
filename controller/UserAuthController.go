package controller

import (
	"strconv"

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

func SelectAllMember(c *fiber.Ctx) error {
	var member []models.Member

	if err := database.DBconn.Find(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถดึงข้อมูลของสมาชิกได้"})
	}

	if len(member) == 0 {
		return c.JSON(fiber.Map{"msg": "ไม่มีสมาชิกในระบบ"})
	}

	return c.JSON(fiber.Map{"msg": member})
}

func Register(c *fiber.Ctx) error {
	var member models.Member

	// ผูก Body ที่ส่งเข้ามาเข้ากับ member
	if err := c.BodyParser(&member); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": err.Error()})
	}

	// สร้าง query ส่งเข้าไปใน database และเก็บไว้ใน existingMember
	var existingMember models.Member
	if err := database.DBconn.Where("email = ?", member.Email).First(&existingMember).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"msg": "มีอีเมลอยู่แล้ว"})
	}

	if err := database.DBconn.Where("phone = ?", member.Phone).First(&existingMember).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"msg": "มีเบอร์โทรอยู่แล้ว"})
	}

	phoneStr := strconv.Itoa(member.Phone)
	if len(phoneStr) != 10 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"msg": "กรุณาใช้เบอร์โทร 10 หลักที่เป็นตัวเลขเท่านั้น"})
	}

	// ทำการ hash รหัสผ่าน
	hashedPassword, err := HashPassword(member.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถแฮชรหัสผ่านได้"})
	}
	// เปลี่ยนรหัสผ่าน
	member.Password = hashedPassword

	// สร้าง user ใหม่
	if err := database.DBconn.Create(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถสร้างผู้ใช้ใหม่ได้"})
	}

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
	if err := database.DBconn.Where("email = ?", input.Email).First(&memberFind).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "เกิดข้อผิดพลาดในการค้นหาผู้ใช้"})
	}

	if memberFind.MemberID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "อีเมลไม่ถูกต้อง"})
	}

	if !CheckPasswordHash(input.Password, memberFind.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "รหัสผ่านไม่ถูกต้อง"})
	}

	return c.JSON(memberFind)
}

func UpdateProfile(c *fiber.Ctx) error {
	mid := c.Params("mid")
	var member models.Member

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone int    `json:"phone"`
	}

	// อ่านข้อมูลจาก body
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "ไม่สามารถแยกวิเคราะห์เนื้อหาคำขอ"})
	}

	// ตรวจสอบข้อมูลที่ส่งมา
	if len(input.Name) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "ชื่อไม่สามารถเว้นว่างได้"})
	}

	if len(input.Email) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "อีเมลไม่สามารถเว้นว่างได้"})
	}

	phoneStr := strconv.Itoa(input.Phone)
	if len(phoneStr) != 10 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "กรุณาใช้เบอร์โทร 10 หลักที่เป็นตัวเลขเท่านั้น"})
	}

	// ค้นหาสมาชิก
	if err := database.DBconn.First(&member, mid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบสมาชิก"})
	}

	// ตรวจสอบการใช้อีเมล
	var existingMember models.Member
	if err := database.DBconn.Where("Email = ? AND MemberID != ?", input.Email, mid).First(&existingMember).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"msg": "อีเมลมีการใช้งานโดยสมาชิกรายอื่นแล้ว"})
	}

	// ตรวจสอบการใช้เบอร์โทร
	if err := database.DBconn.Where("Phone = ? AND MemberID != ?", input.Phone, mid).First(&existingMember).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"msg": "โทรศัพท์มีการใช้งานโดยสมาชิกรายอื่นแล้ว"})
	}

	// อัปเดตข้อมูลสมาชิก
	member.Name = input.Name
	member.Email = input.Email
	member.Phone = input.Phone

	if err := database.DBconn.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถอัปเดตโปรไฟล์ได้"})
	}

	return c.JSON(fiber.Map{"msg": "โปรไฟล์อัปเดตสำเร็จ"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "เกิดข้อผิดพลาดในการแฮชรหัสผ่านใหม่"})
	}

	member.Password = hashedPassword
	if err := database.DBconn.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถบันทึกการเปลี่ยนแปลงรหัสผ่านได้"})
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
