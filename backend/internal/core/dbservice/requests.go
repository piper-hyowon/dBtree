package dbservice

type DBInstanceCreateRequest struct {
	Name        string            `json:"name" validate:"required,min=3,max=15"`
	Type        DBType            `json:"type" validate:"required,one of=mongodb redis"`
	Size        DBSize            `json:"size" validate:"required,one of=small medium large"`
	Mode        DBMode            `json:"mode" validate:"required"`
	Network     NetworkConfig     `json:"network"`
	Backup      BackupConfig      `json:"backup"`
	MongoDBConf *MongoDBConfig    `json:"mongodbConf,omitempty"`
	RedisConf   *RedisConfig      `json:"redisConf,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type DBInstanceUpdateRequest struct {
	Size        *DBSize           `json:"size,omitempty"`
	Backup      *BackupConfig     `json:"backup,omitempty"`
	Network     *NetworkConfig    `json:"network,omitempty"`
	MongoDBConf *MongoDBConfig    `json:"mongodbConf,omitempty"`
	RedisConf   *RedisConfig      `json:"redisConf,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type DBInstanceFilters struct {
	Types    []DBType         `json:"types,omitempty"`
	Sizes    []DBSize         `json:"sizes,omitempty"`
	Statuses []InstanceStatus `json:"statuses,omitempty"`
	NameLike string           `json:"nameLike,omitempty"`
	TagKey   string           `json:"tagKey,omitempty"`
	TagValue string           `json:"tagValue,omitempty"`
}
