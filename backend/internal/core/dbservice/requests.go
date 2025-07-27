package dbservice

import "time"

type CreateInstanceRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	PresetID string `json:"presetId,omitempty"`

	// 커스텀 옵션 (PresetID 없을 때만)
	Type      *DBType                `json:"type,omitempty"`
	Size      *DBSize                `json:"size,omitempty"`
	Mode      *DBMode                `json:"mode,omitempty"`
	Resources *ResourceSpec          `json:"resources,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`

	// 백업 옵션
	BackupEnabled       bool   `json:"backupEnabled"`
	BackupSchedule      string `json:"backupSchedule,omitempty"`
	BackupRetentionDays int    `json:"backupRetentionDays,omitempty"`
}

type UpdateInstanceRequest struct {
	Resources *ResourceSpec          `json:"resources,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`

	// 백업 설정 변경
	BackupEnabled       *bool   `json:"backupEnabled,omitempty"`
	BackupSchedule      *string `json:"backupSchedule,omitempty"`
	BackupRetentionDays *int    `json:"backupRetentionDays,omitempty"`
}

// 목록 조회 필터
type ListInstancesRequest struct {
	Status   *InstanceStatus `json:"status,omitempty"`
	Type     *DBType         `json:"type,omitempty"`
	NameLike string          `json:"nameLike,omitempty"`
	Page     int             `json:"page,omitempty"`
	Limit    int             `json:"limit,omitempty"`
}

type InstanceResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          DBType                 `json:"type"`
	Size          DBSize                 `json:"size"`
	Mode          DBMode                 `json:"mode"`
	Status        InstanceStatus         `json:"status"`
	StatusReason  string                 `json:"statusReason,omitempty"`
	Resources     ResourceSpec           `json:"resources"`
	Cost          CostResponse           `json:"cost"`
	Endpoint      string                 `json:"endpoint,omitempty"`
	Port          int                    `json:"port,omitempty"`
	BackupEnabled bool                   `json:"backupEnabled"`
	Config        map[string]interface{} `json:"config"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

type CostResponse struct {
	CreationCost  int `json:"creationCost"`
	HourlyLemons  int `json:"hourlyLemons"`
	DailyLemons   int `json:"dailyLemons"`
	MonthlyLemons int `json:"monthlyLemons"`
}

type PresetResponse struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Icon        string       `json:"icon"`
	Description string       `json:"description"`
	UseCases    []string     `json:"useCases"`
	Resources   ResourceSpec `json:"resources"`
	Cost        CostResponse `json:"cost"`
}

type ListPresetsResponse struct {
	Redis   []PresetResponse `json:"redis"`
	MongoDB []PresetResponse `json:"mongodb"`
}

// 예상 비용
type EstimateCostRequest struct {
	Type      DBType                 `json:"type"`
	Resources ResourceSpec           `json:"resources"`
	Mode      DBMode                 `json:"mode"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

type EstimateCostResponse struct {
	Cost        CostResponse `json:"cost"`
	Warnings    []string     `json:"warnings,omitempty"`
	Suggestions []string     `json:"suggestions,omitempty"`
}
