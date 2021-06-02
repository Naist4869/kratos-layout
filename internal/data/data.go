package data

import (
	"fmt"

	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	helper  *log.Helper
	mysql   *gorm.DB
	rdb     *redis.Client
	mongodb *mongo.Database
}

func (d *Data) Endpoint() (string, error) {
	return d.mysql.Name(), nil
}

func (d *Data) Start() error {
	return nil
}

func (d *Data) Stop() error {
	db, err := d.mysql.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// NewMysql
func NewMysql(conf *conf.Data, l log.Logger) (db *gorm.DB, cleanup func(), err error) {
	cleanup = func() {
		l.Log(log.LevelInfo, "closing the mysql resources")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
		conf.Mysql.Username,
		conf.Mysql.Password,
		conf.Mysql.Addr,
		conf.Mysql.DbName,
		true,
		//"Asia/Shanghai"),
		"Local")
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return
	}
	db.Set("gorm:table_options", "CHARSET=utf8mb4")
	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	// set for mysql connection
	// 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(int(conf.Mysql.MaxOpenConn))
	// 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(int(conf.Mysql.MaxIdleConn))
	if conf.Mysql.ConnMaxLifeTime != nil {
		sqlDB.SetConnMaxLifetime(conf.Mysql.ConnMaxLifeTime.AsDuration())
	}
	if err = db.AutoMigrate(); err != nil {
		return
	}
	// 创建视图需要exec
	//if db.Find(&[]SysMenu{}).RowsAffected > 0 {
	//	log.Info("\n[Mysql] --> authority_menu 视图已存在!")
	//} else {
	//	if err := db.Exec("CREATE ALGORITHM = UNDEFINED SQL SECURITY DEFINER VIEW `authority_menu` AS select `sys_base_menus`.`id` AS `id`,`sys_base_menus`.`created_at` AS `created_at`, `sys_base_menus`.`updated_at` AS `updated_at`, `sys_base_menus`.`deleted_at` AS `deleted_at`, `sys_base_menus`.`menu_level` AS `menu_level`,`sys_base_menus`.`parent_id` AS `parent_id`,`sys_base_menus`.`path` AS `path`,`sys_base_menus`.`name` AS `name`,`sys_base_menus`.`hidden` AS `hidden`, `sys_base_menus`.`title`  AS `title`,`sys_base_menus`.`icon` AS `icon`,`sys_base_menus`.`sort` AS `sort`,`sys_authority_menus`.`sys_authority_authority_id` AS `authority_id`,`sys_authority_menus`.`sys_base_menu_id` AS `menu_id` from (`sys_authority_menus` join `sys_base_menus` on ((`sys_authority_menus`.`sys_base_menu_id` = `sys_base_menus`.`id`)))").Error; err != nil {
	//		log.Panicf("创建视图表失败: %s, err: %+v", c.Name, err)
	//	}
	//	log.Info("\n[Mysql] --> authority_menu 视图创建成功!")
	//}
	return
}
