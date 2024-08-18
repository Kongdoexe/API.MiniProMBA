package database

import (
	"fmt"
	"log"
	"os"

	"github.com/Kongdoexe/goland/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBconn *gorm.DB

func InitDB() {

	user := os.Getenv("db_user")
	pass := os.Getenv("db_password")
	host := os.Getenv("db_host")
	dbname := os.Getenv("db_dbname")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})

	if err != nil {
		panic("Database connect failed.")
	}

	log.Println("Connect Success.")

	db.AutoMigrate(&models.Member{}, &models.LottoTicket{}, &models.RankLotto{}, &models.Winner{}, &models.Cart{})
	// db.AutoMigrate(&models.LottoTicket{})

	// db.AutoMigrate(&models.Member{}, &models.LottoTicket{}, &models.Cart{}, &models.RankLotto{}, &models.Winner{})

	DBconn = db
}
