package resource

const (
	// EC2 m5a.large

	TotalCPU    = 2.0
	TotalMemory = 8192 // 8GB in MB

	SystemReservedCPU    = 0.5
	SystemReservedMemory = 1536 // 1.5GB in MB

	AvailableCPU    = TotalCPU - SystemReservedCPU       // 1.5
	AvailableMemory = TotalMemory - SystemReservedMemory // 6656 MB

	// 프리셋별 리소스 (참조용, TODO: DB랑 불일치 주의..)

	TinyCPU      = 0.1
	TinyMemory   = 256
	SmallCPU     = 0.25
	SmallMemory  = 512
	MediumCPU    = 0.5
	MediumMemory = 1024
)
