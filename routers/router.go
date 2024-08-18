package routers

import (
	"github.com/Kongdoexe/goland/controller"
	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App) {

	Auth := app.Group("/auth")
	Lottery := app.Group("/lottery")
	Wallet := app.Group("/wallet")

	Sales := app.Group("/sales")
	Draw := app.Group("/draw")

	Auth.Get("/Login", controller.Login)                         //ล็อคอิน -Pa
	Auth.Post("/Register", controller.Register)                  //สมัครสมาชิก -Pa
	Auth.Put("/UpdateProfile/:mid", controller.UpdateProfile)    //แก้ไขข้อมูลส่วนตัว -Pa
	Auth.Put("/ChangePassword/:mid", controller.ChangePassword)  //เปลี่ยนรหัสผ่าน -Pa
	Auth.Delete("/DeleteAccount/:mid", controller.DeleteAccount) //ลบบัญชี -Pa

	Wallet.Put("/AddFunds/:mid", controller.AddFunds) //เพิ่มจำนวนเงิน -Pa

	Lottery.Post("/insertLottery", controller.InsertLotto)                        //เพิ่มลอตโต้ -Pa
	Lottery.Get("/GetDrawSchedule", controller.GetDrawSchedule)                   //ค้นหารอบของทุกเลข -Pa
	Lottery.Post("/InsertCart", controller.InsertCart)                            //เพิ่มลงตระกร้า -Pa
	Lottery.Get("/GetNumbersInCart/:mid", controller.GetNumbersInCart)            //ค้นหาเลขที่อยู่ในตระกร้าของผู้ใช้ -Pa
	Lottery.Delete("/RemoveNumberFromCart/:cid", controller.RemoveNumberFromCart) //ลบเลขที่ผู้ใช้เอาออกจากตระกร้า -Pa
	Lottery.Put("/ProcessPayment/:mid", controller.ProcessPayment)                //ผู้ใช้ชำระเงิน -Pa
	Lottery.Get("/SelectDraw/:period", controller.SelectDraw)                     //เลือกรอบที่ออกรางวัล -Pa
	Lottery.Get("/GetUserDrawNumbers/:mid", controller.GetUserDrawNumbers)        //เอาเลขของผู้ใช้ในรอบปัจจุบันออกมา -Pa
	Lottery.Get("/GetWinningNumbers", controller.GetWinningNumbers)               //ค้นหาเลขที่ผู้ใช้ถูกรางวัล -Pa
	Lottery.Put("/AddWinningsToWallet", controller.AddWinningsToWallet)           //เพิ่มจำนวนเงินผู้ใช้ตามที่ถูกรางวัล -Pa

	Sales.Get("/GetSalesData", controller.GetSalesData) //รวมข้อมูลยอดขาย -Pa
	// Sales.Get("/GetSalesSummary", controller.GetSalesSummary)                   //สรุปยอดขาย
	// Sales.Get("/GetSoldNumbersCount", controller.GetSoldNumbersCount)           //จำนวนที่ขายได้
	// Sales.Get("/GetRemainingNumbersCount", controller.GetRemainingNumbersCount) //จำนวนคงเหลือ

	Draw.Post("/GenerateUniqueDraw", controller.GenerateUniqueDraw) //สุ่มออกรางวัล โดยเลือกวิธีออกรางวัลได้ -Pa
	Draw.Get("/Getprize", controller.Getprize)                      //เอาลอตโต้ที่ชนะในแต่ละรอบ -Pa
	Draw.Delete("/ResetSystem", controller.ResetSystem)             //รีเซ็ทระบบ -Pa

}
