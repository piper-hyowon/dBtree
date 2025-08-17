package dbservice

import (
	"github.com/piper-hyowon/dBtree/internal/core/errors"
	"time"
)

type CreateInstanceRequest struct {
	Name     string  `json:"name" validate:"required,min=3,max=63,instancename"`
	PresetID *string `json:"presetId,omitempty"`

	// 커스텀 옵션 (PresetID 없을 때만)
	Type      *DBType                `json:"type,omitempty"`
	Mode      *DBMode                `json:"mode,omitempty"`
	Resources *ResourceSpec          `json:"resources,omitempty" validate:"omitempty,dive"`
	Config    map[string]interface{} `json:"config,omitempty"`

	// 백업 옵션
	BackupEnabled       bool   `json:"backupEnabled"`
	BackupSchedule      string `json:"backupSchedule,omitempty" validate:"omitempty,cronschedule"`
	BackupRetentionDays int    `json:"backupRetentionDays,omitempty" validate:"min=0,max=365"`
}

type CreateInstanceResponse struct {
	ID          string       `json:"id"` // ExternalID
	Name        string       `json:"name"`
	Type        DBType       `json:"type"`
	Status      string       `json:"status"`
	Resources   ResourceSpec `json:"resources"`
	Cost        LemonCost    `json:"cost"`
	CreatedAt   time.Time    `json:"createdAt"`
	Credentials *Credentials `json:"credentials,omitempty"` // 생성시에만 포함
}

type Credentials struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ExternalHost string `json:"externalHost,omitempty"`
	ExternalPort int    `json:"externalPort,omitempty"`
	ExternalURI  string `json:"externalUri,omitempty"`
}

type UpdateInstanceRequest struct {
	Resources *ResourceSpec          `json:"resources,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`

	// 백업 설정 변경
	BackupEnabled       *bool   `json:"backupEnabled,omitempty"`
	BackupSchedule      *string `json:"backupSchedule,omitempty"`
	BackupRetentionDays *int    `json:"backupRetentionDays,omitempty"`
}

type InstanceResponse struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Type                DBType                 `json:"type"`
	Size                DBSize                 `json:"size"`
	Mode                DBMode                 `json:"mode"`
	Status              InstanceStatus         `json:"status"`
	StatusReason        string                 `json:"statusReason,omitempty"`
	Resources           ResourceSpec           `json:"resources"`
	Cost                CostResponse           `json:"cost"`
	Endpoint            string                 `json:"endpoint,omitempty"`
	Port                int                    `json:"port,omitempty"`
	ExternalHost        string                 `json:"externalHost,omitempty"`
	ExternalPort        int                    `json:"externalPort,omitempty"`
	ExternalURITemplate string                 `json:"externalUriTemplate,omitempty"`
	BackupEnabled       bool                   `json:"backupEnabled"`
	Config              map[string]interface{} `json:"config"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`
	CreatedFromPreset   *string                `json:"createdFromPreset,omitempty"`
	PausedAt            *time.Time             `json:"pausedAt,omitempty"`
}

type CostResponse struct {
	CreationCost  int `json:"creationCost"`
	HourlyLemons  int `json:"hourlyLemons"`
	DailyLemons   int `json:"dailyLemons"`
	MonthlyLemons int `json:"monthlyLemons"`
}

type PresetResponse struct {
	ID                  string                 `json:"id"`
	Type                DBType                 `json:"type"`
	Size                DBSize                 `json:"size"`
	Mode                DBMode                 `json:"mode"`
	Name                string                 `json:"name"`
	Icon                string                 `json:"icon"`
	Description         string                 `json:"description"`
	FriendlyDescription string                 `json:"friendlyDescription"`
	TechnicalTerms      map[string]interface{} `json:"technicalTerms,omitempty"`
	UseCases            []string               `json:"useCases"`
	Resources           ResourceSpec           `json:"resources"`
	Cost                CostResponse           `json:"cost"`
	DefaultConfig       map[string]interface{} `json:"defaultConfig,omitempty"`
	SortOrder           int                    `json:"sortOrder"`
	Available           bool                   `json:"available"`
	UnavailableReason   string                 `json:"unavailableReason,omitempty"`
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

func (r *CreateInstanceRequest) Validate() error {
	// PresetID와 커스텀 옵션은 동시에 사용할 수 없음
	if r.PresetID != nil && (r.Type != nil || r.Resources != nil) {
		return errors.NewInvalidParameterError("request",
			"PresetID와 커스텀 옵션(type, resources)은 동시에 사용할 수 없습니다")
	}

	// PresetID 없으면 type과 resources 필수
	if r.PresetID == nil && (r.Type == nil || r.Resources == nil) {
		return errors.NewInvalidParameterError("request",
			"PresetID가 없으면 type과 resources를 반드시 지정해야 합니다")
	}

	// DBType 유효성 검증
	if r.Type != nil {
		// 현재는 MongoDB만 지원
		if *r.Type == Redis {
			return errors.NewInvalidParameterError("type", "Redis는 아직 지원하지 않습니다")
		}
		if *r.Type != MongoDB {
			return errors.NewInvalidParameterError("type", "지원하지 않는 데이터베이스 타입입니다")
		}
	}

	// DBMode 유효성 검증 - Type에 따라 다른 모드 허용
	if r.Mode != nil && r.Type != nil {
		switch *r.Type {
		case MongoDB:
			validModes := map[DBMode]bool{
				ModeStandalone: true,
				ModeReplicaSet: true,
				ModeSharded:    true,
			}
			if !validModes[*r.Mode] {
				return errors.NewInvalidParameterError("mode",
					"MongoDB는 standalone, replica_set, sharded 모드만 지원합니다")
			}
		case Redis:
			validModes := map[DBMode]bool{
				ModeBasic:    true,
				ModeSentinel: true,
				ModeCluster:  true,
			}
			if !validModes[*r.Mode] {
				return errors.NewInvalidParameterError("mode",
					"Redis는 basic, sentinel, cluster 모드만 지원합니다")
			}
		}
	}

	// Mode가 없으면 기본값 설정
	if r.Type != nil && r.Mode == nil {
		defaultMode := r.Type.DefaultMode()
		r.Mode = &defaultMode
	}

	// 백업 설정 검증
	if r.BackupEnabled {
		// 백업이 활성화되면 스케줄 필수
		if r.BackupSchedule == "" {
			return errors.NewInvalidParameterError("backupSchedule",
				"백업이 활성화되면 백업 스케줄을 지정해야 합니다")
		}

		// RetentionDays 기본값 설정
		if r.BackupRetentionDays == 0 {
			r.BackupRetentionDays = 7
		}
	}

	return nil
}
