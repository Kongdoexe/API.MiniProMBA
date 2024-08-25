package main

import (
	"log"

	"github.com/Kongdoexe/goland/database"
	"github.com/Kongdoexe/goland/routers"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func init() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error in loading .env file.")
	}

	database.InitDB()
}

func main() {
	app := fiber.New()

	sqlDb, err := database.DBconn.DB()

	if err != nil {
		panic("Error in sql connection.")
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"msg": "Hello, World!!!"})
	})

	defer sqlDb.Close()

	routers.SetupRouter(app)

	app.Listen(":3000")
}
