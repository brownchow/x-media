package databases

import (
	"database/sql"
	"fmt"

	"github.com/0x113/x-media/user/common"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

// MysqlDB maanages MySQL connection
type MysqlDB struct {
	DB     *sql.DB
	DbName string
}

// Init initializes MySQL database
func (database *MysqlDB) Init() error {
	log.Infoln("Connecting to the MySQL datbase...")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", common.Config.DbUsername, common.Config.DbPassword, common.Config.DbAddr, common.Config.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// check connection
	if err := db.Ping(); err != nil {
		return err
	}

	// set db
	database.DB = db
	database.DbName = common.Config.DbName
	return nil
}
