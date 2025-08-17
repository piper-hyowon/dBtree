package resource

// SystemResources 시스템 전체 리소스 정보
type SystemResources struct {
	Total     SystemResourceSpec `json:"total"`
	Reserved  SystemResourceSpec `json:"reserved"` // K3s + System
	Available SystemResourceSpec `json:"available"`
	Used      SystemResourceSpec `json:"used"`
}

type SystemResourceSpec struct {
	CPU    float64 `json:"cpu"`    // vCPU
	Memory int     `json:"memory"` // MB
}

type InstanceResources struct {
	InstanceID   string             `json:"instanceId"`
	InstanceName string             `json:"instanceName"`
	Resources    SystemResourceSpec `json:"resources"`
	Status       string             `json:"status"`
}

type SystemResourceStatus struct {
	Info            SystemResources     `json:"system"`
	Instances       []InstanceResources `json:"instances"`
	ActiveCount     int                 `json:"activeCount"`
	CanCreateTiny   bool                `json:"canCreateTiny"`   // Tiny 생성 가능 여부
	CanCreateSmall  bool                `json:"canCreateSmall"`  // Small 생성 가능 여부
	CanCreateMedium bool                `json:"canCreateMedium"` // Medium 생성 가능 여부
}
