package db

import (
	"errors"
	"github.com/JaanaiShi/flint/common"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// @description: 初始化Mysql数据库
// @return: *gorm.DB
func Init() {

	m := common.Conf.DB
	if m.Name == "" {
		panic(errors.New("MySQL用户名配置为空"))
	}
	dsn := m.Username + ":" + m.Password + "@tcp(" + m.Host + ":" + m.Port + ")/" + m.Name + "?" + m.Config
	mysqlConfig := mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         191,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置

	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), gormConfig(m.DbLogMode)); err != nil {
		panic(err)
	} else {
		sqlDB, err := db.DB()
		if err != nil {
			panic(err)
		}
		sqlDB.SetMaxIdleConns(m.MaxIdle)
		sqlDB.SetMaxOpenConns(m.MaxOpen)
		common.DB = db
		return
	}
}

// @description: 根据配置决定是否开启日志
func gormConfig(mod bool) *gorm.Config {
	c := common.Conf
	var config = &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}
	if !mod {
		return config
	}
	switch c.DB.LogZap {
	case "silent", "Silent":
		config.Logger = Default.LogMode(gormLogger.Silent)
	case "error", "Error":
		config.Logger = Default.LogMode(gormLogger.Error)
	case "warn", "Warn":
		config.Logger = Default.LogMode(gormLogger.Warn)
	case "info", "Info":
		config.Logger = Default.LogMode(gormLogger.Info)
	case "zap", "Zap":
		config.Logger = Default.LogMode(gormLogger.Info)
	default:
		if mod {
			config.Logger = Default.LogMode(gormLogger.Info)
			break
		}
		config.Logger = Default.LogMode(gormLogger.Silent)
	}
	return config
}
