package version

// DbMeta 数据库元数据信息
type DbMeta struct {
	ID      int64 `bson:"-" json:"-"`
	Version int64 `bson:"version"` // 版本
}
