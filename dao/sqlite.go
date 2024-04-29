package dao

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"h-ui/model/constant"
	"log"
	"os"
	"strings"
	"time"
)

var sqliteDB *gorm.DB

var sqlInitStr = "create table account\n(\n    id          INTEGER                  not null\n        primary key autoincrement,\n    username    TEXT      default ''     not null,\n    `pass`      TEXT      default ''     not null,\n    con_pass    TEXT      default ''     not null,\n    quota       INTEGER   default 0      not null,\n    download    INTEGER   default 0      not null,\n    upload      INTEGER   default 0      not null,\n    expire_time INTEGER   default 0      not null,\n    role        TEXT      default 'user' not null,\n    deleted     INTEGER   default 0      not null,\n    create_time TIMESTAMP default CURRENT_TIMESTAMP,\n    update_time TIMESTAMP default CURRENT_TIMESTAMP\n);\n\ncreate index account_deleted_index\n    on account (deleted);\n\ncreate index account_con_pass_index\n    on account (con_pass);\n\ncreate index account_pass_index\n    on account (`pass`);\n\ncreate index account_username_index\n    on account (username);\n\nINSERT INTO account (username, `pass`, con_pass, quota, download, upload, expire_time, role)\nVALUES ('sysadmin', 'f8cdb04495ded47615258f9dc6a3f4707fd2405434fefc3cbf4ef4e6',\n        'c7591c31adf8af0b6b8ae8cbbccd8d1aaa0c7bb068f576bddb6378d5', -1, 0, 0, 253370736000000, 'admin');\n\n\ncreate table config\n(\n    id          INTEGER              not null\n        primary key autoincrement,\n    `key`       TEXT      default '' not null,\n    `value`     TEXT      default '' not null,\n    remark      TEXT      default '' not null,\n    create_time TIMESTAMP default CURRENT_TIMESTAMP,\n    update_time TIMESTAMP default CURRENT_TIMESTAMP\n);\n\ncreate index account_key_index\n    on config (`key`);\n\nINSERT INTO config (key, value, remark)\nVALUES ('H_UI_WEB_PORT', '8081', 'H UI Web 端口');\nINSERT INTO config (key, value, remark)\nVALUES ('JWT_SECRET', hex(randomblob(10)), 'JWT 密钥');\nINSERT INTO config (key, value, remark)\nVALUES ('HYSTERIA2_ENABLE', '0', 'Hysteria2 开关');\nINSERT INTO config (key, value, remark)\nVALUES ('HYSTERIA2_CONFIG', '', 'Hysteria2 配置');\nINSERT INTO config (key, value, remark)\nVALUES ('HYSTERIA2_TRAFFIC_TIME', '1', 'Hysteria2 流量倍数');"

func InitSqliteDB() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
	)

	var err error
	sqliteDB, err = gorm.Open(sqlite.Open(constant.SqliteDBPath), &gorm.Config{
		TranslateError: true,
		Logger:         newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("sqlite connect err: %v", err))
	}

	var count uint
	if err := sqliteDB.Raw("SELECT count(1) FROM sqlite_master WHERE type='table' AND (name = 'account' or name = 'config')").Scan(&count).Error; err != nil {
		logrus.Errorf("sqlite query database err: %v", err)
		panic(err)
	}
	if count == 0 {
		if err = sqliteInit(sqlInitStr); err != nil {
			logrus.Errorf("sqlite database import err: %v", err)
			panic(err)
		}
	}
}

func sqliteInit(sqlStr string) error {
	if sqliteDB != nil {
		sqls := strings.Split(strings.Replace(sqlStr, "\r\n", "\n", -1), ";\n")
		for _, s := range sqls {
			s = strings.TrimSpace(s)
			if s != "" {
				tx := sqliteDB.Exec(s)
				if tx.Error != nil {
					logrus.Errorf("sqlite exec err: %v", tx.Error.Error())
					panic(tx.Error.Error())
				}
			}
		}
	}
	return nil
}

func CloseSqliteDB() {
	if sqliteDB != nil {
		db, err := sqliteDB.DB()
		if err != nil {
			logrus.Errorf("sqlite err: %v", err)
			return
		}
		if err = db.Close(); err != nil {
			logrus.Errorf("sqlite close err: %v", err)
		}
	}
}

func Paginate(pageNumParam *int64, pageSizeParam *int64) func(db *gorm.DB) *gorm.DB {
	var pageNum int64 = 1
	var pageSize int64 = 10
	if pageNumParam == nil || *pageNumParam == 0 {
		pageNum = *pageNumParam
	}
	if pageSizeParam == nil || *pageSizeParam == 0 {
		pageSize = *pageSizeParam
	}
	return func(db *gorm.DB) *gorm.DB {
		offset := (pageNum - 1) * pageSize
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}