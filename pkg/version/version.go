package version

// DbMeta 数据库元数据信息
type DbMeta struct {
	Version int `bson:"version"` // 版本
}
