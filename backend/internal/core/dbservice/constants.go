package dbservice

const (
	// MaxInstancesPerUser 사용자당 최대 인스턴스 개수
	MaxInstancesPerUser = 2
)

// ReservedPorts 포트 할당 할때 건너뜀
var ReservedPorts = map[int]bool{
	30080: true, // backend
}
