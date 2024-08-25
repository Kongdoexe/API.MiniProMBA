package controller

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/Kongdoexe/goland/database"
	"github.com/Kongdoexe/goland/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type WinningsToWallet struct {
	ID   int64
	Data []Result
}

type Result struct {
	HasWinner bool
	Number    string
	Gratuity  int
}

var (
	generatedNumbers = make(map[string]bool)
	mu               sync.Mutex // ใช้สำหรับการทำงานพร้อมกันอย่างปลอดภัย
)

// ประกาศ seed สำหรับการสุ่มเพียงครั้งเดียว
func init() {
	rand.Seed(time.Now().UnixNano())
}

// ฟังก์ชันสำหรับการสร้างหมายเลขล็อตโต้ที่ไม่ซ้ำ
func generateUniqueLotto() string {
	mu.Lock()
	defer mu.Unlock()

	for {
		lottoNumber := ""
		for i := 0; i < 6; i++ {
			lottoNumber += strconv.Itoa(rand.Intn(10))
		}

		if !generatedNumbers[lottoNumber] {
			generatedNumbers[lottoNumber] = true
			return lottoNumber
		}
	}
}

// ฟังก์ชันสำหรับการสร้าง 100 หมายเลขล็อตโต้ที่ไม่ซ้ำ
func generate100UniqueLottos() []string {
	lottoNumbers := []string{}
	uniqueNumbers := make(map[string]bool)

	for len(lottoNumbers) < 100 {
		number := fmt.Sprintf("%06d", rand.Intn(1000000))
		if !uniqueNumbers[number] {
			lottoNumbers = append(lottoNumbers, number)
			uniqueNumbers[number] = true
		}
	}

	return lottoNumbers
}

func SelectAllLotto(c *fiber.Ctx) error {
	var lotto []models.LottoTicket

	if err := database.DBconn.Find(&lotto).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถดึงข้อมูลออกมาดูได้"})
	}

	if len(lotto) == 0 {
		return c.JSON(fiber.Map{"msg": "ไม่มีข้อมูลลอตโต้"})
	}

	return c.JSON(lotto)
}

func InsertLotto(c *fiber.Ctx) error {
	lottoNumbers := generate100UniqueLottos()
	var latestLotto models.LottoTicket

	if err := database.DBconn.Last(&latestLotto).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			latestLotto.Period = 1
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "ไม่สามารถดึงข้อมูลลอตโต้ล่าสุดได้",
			})
		}
	} else {
		var count int64
		err := database.DBconn.Table("Winner").
			Joins("JOIN LottoTicket ON Winner.TicketID = LottoTicket.TicketID").
			Where("LottoTicket.Period = ?", latestLotto.Period).
			Count(&count).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "ไม่สามารถดึงข้อมูลรางวัลที่ออกได้",
			})
		}

		if count >= 5 {
			latestLotto.Period += 1
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"msg": "ลอตโต้ชุดนี้เริ่มออกรางวัลแล้ว ไม่สามารถเพิ่มได้",
			})
		}
	}

	for _, number := range lottoNumbers {
		datainsert := models.LottoTicket{
			MemberID: nil,
			Number:   number,
			Period:   latestLotto.Period,
			Price:    80,
		}

		if err := database.DBconn.Create(&datainsert).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "ไม่สามารถเพิ่มลอตโต้ได้",
			})
		}
	}

	return c.JSON(fiber.Map{"msg": "เพิ่มลอตโต้เรียบร้อย"})
}

func GetDrawSchedule(c *fiber.Ctx) error {
	var lotto []models.LottoTicket
	var selectPeriod models.LottoTicket

	if err := database.DBconn.Order("Period DESC").First(&selectPeriod).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "ไม่สามารถดึงข้อมูลช่วงเวลาล่าสุดได้",
		})
	}

	if err := database.DBconn.Where("Period = ? AND MemberID IS NULL", selectPeriod.Period).Find(&lotto).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "ไม่สามารถเรียกตารางการออกรางวัลได้",
		})
	}

	if len(lotto) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"msg": "ไม่พบลอตโต้สำหรับช่วงเวลานี้ หรือ ลอตเตอรี่ไม่ได้ถูกเพิ่ม",
		})
	}

	return c.JSON(lotto)
}

func InsertCart(c *fiber.Ctx) error {
	var Lotto models.LottoTicket
	var Member models.Member
	var Cart models.Cart
	var data struct {
		TicketID int `json:"ticketID"`
		MemberID int `json:"memberID"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": err.Error()})
	}

	if err := database.DBconn.Where("TicketID = ? AND MemberID IS NULL", data.TicketID).First(&Lotto).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบล็อตโต้หรือถูกซื้อไปแล้วไปแล้ว"})
	}

	result := database.DBconn.Where("LottoTicketID = ? AND MemberID = ?", data.TicketID, data.MemberID).First(&Cart)
	if result.RowsAffected > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "มีลอตโต้เลขนี้ในตระกร้าแล้ว"})
	}

	var count int64
	err := database.DBconn.Table("Winner").
		Joins("JOIN LottoTicket ON Winner.TicketID = LottoTicket.TicketID").
		Where("LottoTicket.Period = ?", Lotto.Period).
		Count(&count).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถดึงข้อมูลรางวัลที่ออกได้"})
	}

	if count >= 1 {
		return c.JSON(fiber.Map{"msg": "ปิดการซื้อขายแล้วในรอบนี้"})
	}

	if err := database.DBconn.First(&Member, data.MemberID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบผู้ใช้"})
	}

	Cart.LottoTicketID = data.TicketID
	Cart.MemberID = data.MemberID

	if err := database.DBconn.Create(&Cart).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถสร้างตระกร้าได้"})
	}

	return c.JSON(fiber.Map{"msg": "เพิ่มลอตโต้ลงตระกร้าเรียบร้อย"})
}

func GetNumbersInCart(c *fiber.Ctx) error {
	mid := c.Params("mid")
	var Cart []models.Cart
	var Member models.Member
	var response []struct {
		CartID   int    `json:"cart_id"`
		MemberID int    `json:"member_id"`
		NumLotto string `json:"num_lotto"`
	}

	if err := database.DBconn.First(&Member, mid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบผู้ใช้"})
	}

	if err := database.DBconn.Where("MemberID = ?", mid).Find(&Cart).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "ไม่สามารถเรียกข้อมูลสินค้าในตระกร้าได้",
		})
	}

	if len(Cart) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่มีตั๋วล็อตโต้ในตระกร้า"})
	}

	for _, ticket := range Cart {
		var lottoTicket models.LottoTicket
		if err := database.DBconn.Where("TicketID = ?", ticket.LottoTicketID).First(&lottoTicket).Error; err == nil {
			response = append(response, struct {
				CartID   int    `json:"cart_id"`
				MemberID int    `json:"member_id"`
				NumLotto string `json:"num_lotto"`
			}{
				CartID:   ticket.CartID,
				MemberID: ticket.MemberID,
				NumLotto: lottoTicket.Number,
			})
		}
	}

	return c.JSON(response)
}

func RemoveNumberFromCart(c *fiber.Ctx) error {
	var data struct {
		CartID   int `json:"cartID"`
		MemberID int `json:"memberID"`
	}

	// อ่านข้อมูลจาก body
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "ไม่สามารถแยกวิเคราะห์ข้อมูลคำขอได้"})
	}

	// ตรวจสอบความถูกต้องของข้อมูล
	if data.CartID == 0 || data.MemberID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "รหัสตระกร้าหรือรหัสสมาชิกไม่ถูกต้อง"})
	}

	// ลบข้อมูลจากฐานข้อมูล
	result := database.DBconn.Delete(&models.Cart{}, "CartID = ? AND MemberID = ?", data.CartID, data.MemberID)

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบตระกร้าที่ตรงกับรหัสที่ระบุ"})
	}

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถลบข้อมูลตระกร้าได้"})
	}

	return c.JSON(fiber.Map{"msg": "ลบตระกร้าสำเร็จ"})
}

func ProcessPayment(c *fiber.Ctx) error {
	mid := c.Params("mid")
	var member models.Member
	var cart []models.Cart
	var lottodelete []models.LottoTicket
	var cashpay int

	// ตรวจสอบว่าพบ Member หรือไม่
	if err := database.DBconn.First(&member, mid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่พบผู้ใช้"})
	}

	// ดึงข้อมูล Cart ของ Member
	if err := database.DBconn.Where("MemberID = ?", mid).Find(&cart).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "ไม่สามารถเรียกข้อมูลตะกร้าได้",
		})
	}

	// ตรวจสอบว่ามีรายการใน Cart หรือไม่
	if len(cart) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ไม่มีล็อตโต้ในตะกร้า"})
	}

	successfulPurchases := 0

	for _, cartItem := range cart {
		var lottoTicket models.LottoTicket

		// ค้นหาหมายเลขล็อตโต้ในฐานข้อมูล
		if err := database.DBconn.Where("TicketID = ?", cartItem.LottoTicketID).First(&lottoTicket).Error; err != nil {
			// ถ้าหมายเลขล็อตโต้ไม่พบ ให้ลบรายการจากตะกร้าและดำเนินการต่อ
			if err := database.DBconn.Delete(&cartItem).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "เกิดข้อผิดพลาดในการลบลอตโต้ที่ไม่พบ"})
			}
			continue
		}

		if lottoTicket.MemberID != nil {
			// เพิ่มล็อตโต้ที่ไม่สามารถซื้อได้ไปยังรายการที่ต้องลบ
			lottodelete = append(lottodelete, lottoTicket)
			if err := database.DBconn.Delete(&cartItem).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "เกิดข้อผิดพลาดในการลบลอตโต้ที่ถูกซื้อไปแล้ว"})
			}
			continue
		}

		// อัพเดตล็อตโต้ให้เป็นของสมาชิกปัจจุบัน
		lottoTicket.MemberID = &member.MemberID
		if err := database.DBconn.Save(&lottoTicket).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถบันทึกข้อมูลล็อตโต้ได้"})
		}

		successfulPurchases++
	}

	cashpay = successfulPurchases * 80

	if member.WalletBalance < cashpay {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "ยอดเงินในกระเป๋าไม่เพียงพอ"})
	}

	// หักเงินจากกระเป๋าของสมาชิก
	member.WalletBalance -= cashpay

	if err := database.DBconn.Save(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ล้มเหลวในการอัพเดตยอดเงินในกระเป๋า"})
	}

	// ลบรายการทั้งหมดจาก Cart หลังจากทำการซื้อสำเร็จ
	if err := database.DBconn.Delete(&cart).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"msg": "ไม่สามารถลบตระกร้าหลังจากการซื้อสำเร็จ",
		})
	}

	return c.JSON(fiber.Map{
		"msg":         "ดำเนินการชำระเงินเรียบร้อยแล้ว",
		"LottoDelete": lottodelete,
	})
}

func SelectDraw(c *fiber.Ctx) error {
	period := c.Params("period")
	var Winner []struct {
		WinnerID uint
		TicketID uint
		MemberID uint
		Number   string
		Rank     uint
	}

	result := database.DBconn.Table("Winner").
		Select("Winner.WinnerID, LottoTicket.TicketID, LottoTicket.MemberID, LottoTicket.Number, Winner.Rank").
		Joins("JOIN LottoTicket ON LottoTicket.TicketID = Winner.TicketID").
		Where("LottoTicket.Period = ?", period).
		Find(&Winner).Error

	if result != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "เกิดข้อผิดพลาดในการดึงข้อมูลรอบรางวัล"})
	}

	if len(Winner) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "ยังไม่ออกรอบรางวัลในรอบนี้"})
	}

	return c.JSON(Winner)
}

func GetUserDrawNumbers(c *fiber.Ctx) error {
	mid := c.Params("mid")
	var selectPeriod models.LottoTicket

	if err := database.DBconn.Order("Period DESC").First(&selectPeriod).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถเรียกข้อมูลช่วงเวลาล่าสุดได้"})
	}

	var resultLotto []models.LottoTicket
	result := database.DBconn.Table("LottoTicket").
		Where("LottoTicket.MemberID = ? AND LottoTicket.Period = ?", mid, selectPeriod.Period).
		Order("LottoTicket.TicketID ASC").
		Find(&resultLotto).Error

	if len(resultLotto) == 0 {
		return c.JSON(fiber.Map{"msg": "คุณยังไม่ซื้อลอตโต้ในรอบนี้"})
	}

	if result != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "เกิดข้อผิดพลาดในการดึงข้อมูลลอตโต้ของผู้ใช้"})
	}

	return c.JSON(resultLotto)
}

func GetWinningNumbers(c *fiber.Ctx) error {
	var requestWinnerCheck []struct {
		Number string `json:"number"`
		Period int    `json:"period"`
	}

	if err := c.BodyParser(&requestWinnerCheck); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": err.Error(),
		})
	}

	results := make([]Result, len(requestWinnerCheck))

	for i, check := range requestWinnerCheck {
		result, err := CheckWinner(database.DBconn, check.Number, check.Period)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"msg": "ไม่สามารถตรวจสอบผู้ชนะได้",
			})
		}
		results[i] = result
	}

	return c.JSON(results)
}

func CheckWinner(db *gorm.DB, number string, period int) (Result, error) {

	var result Result

	query := `
    SELECT 
        CASE 
            WHEN COUNT(*) > 0 THEN 'true' 
            ELSE 'false'
        END AS HasWinner,
        ? AS Number,
		RankLotto.Gratuity AS Gratuity
    FROM Winner
    JOIN LottoTicket ON Winner.TicketID = LottoTicket.TicketID
    JOIN RankLotto ON RankLotto.RankID = Winner.Rank
    WHERE LottoTicket.Number = ?
    AND LottoTicket.Period = ?
    `

	err := db.Raw(query, number, number, period).Scan(&result).Error
	if err != nil {
		return Result{}, err
	}

	return result, nil
}

func AddWinningsToWallet(c *fiber.Ctx) error {
	var form WinningsToWallet
	var member models.Member

	if err := c.BodyParser(&form); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": err.Error(),
		})
	}

	member.MemberID = int(form.ID)
	if err := database.DBconn.First(&member).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"msg": "ไม่สามารถดึงข้อมูลจากผู้ใช้ได้",
		})
	}

	for _, data := range form.Data {
		if data.HasWinner {
			member.WalletBalance += data.Gratuity

			if err := database.DBconn.Save(&member).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "ไม่สามารถอัปเดตยอดเงินในกระเป๋าได้"})
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"msg": "เพิ่มเงินรางวัลเข้ากระเป๋าเงินเรียบร้อยแล้ว",
	})
}
