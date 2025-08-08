package stats

import (
	"github.com/piper-hyowon/dBtree/internal/common"
	"github.com/piper-hyowon/dBtree/internal/core/lemon"
)

type DailyHarvestRequest struct {
	Days int `json:"days" validate:"min=1,max=365"`
}

type TransactionsRequest struct {
	common.PaginationParams
	InstanceName *string `json:"instanceName,omitempty"`
}

type TransactionsResponse struct {
	Data       []*lemon.TransactionWithInstance `json:"data"`
	Pagination *common.PaginationInfo           `json:"pagination"`
}
