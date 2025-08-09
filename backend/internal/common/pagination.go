package common

type PaginationParams struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=100"`
}

// GetOffset 현재 페이지를 기반 오프셋 계산
func (p *PaginationParams) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

// SetDefaults 기본값 설정 (각 패키지에서 자체적으로 호출)
func (p *PaginationParams) SetDefaults(defaultLimit int) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = defaultLimit
	}
}

type PaginationInfo struct {
	CurrentPage int  `json:"currentPage"`
	TotalPages  int  `json:"totalPages"`
	TotalItems  int  `json:"totalItems"`
	HasNext     bool `json:"hasNext"`
	HasPrev     bool `json:"hasPrev"`
}

func NewPaginationInfo(currentPage, limit, totalItems int) *PaginationInfo {
	totalPages := (totalItems + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &PaginationInfo{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		HasNext:     currentPage < totalPages,
		HasPrev:     currentPage > 1,
	}
}
