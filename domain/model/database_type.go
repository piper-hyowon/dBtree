package model

type DatabaseType string

const (
	MongoDB DatabaseType = "mongodb"
	Redis   DatabaseType = "redis"
)

var DefaultResourceSpecs = map[DatabaseType]map[string]ResourceSpec{
	MongoDB: {
		"small":  {CPU: 1, Memory: 1024, Disk: 10},
		"medium": {CPU: 2, Memory: 2048, Disk: 20},
		"large":  {CPU: 4, Memory: 4096, Disk: 40},
	},
	Redis: {
		"small":  {CPU: 1, Memory: 512, Disk: 5},
		"medium": {CPU: 2, Memory: 1024, Disk: 10},
		"large":  {CPU: 2, Memory: 2048, Disk: 20},
	},
}
