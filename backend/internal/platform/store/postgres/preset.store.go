package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/piper-hyowon/dBtree/internal/core/dbservice"
	"github.com/piper-hyowon/dBtree/internal/core/errors"
)

type PresetStore struct {
	db *sql.DB
}

var _ dbservice.PresetStore = (*PresetStore)(nil)

func NewPresetStore(db *sql.DB) dbservice.PresetStore {
	return &PresetStore{
		db: db,
	}
}

func (s *PresetStore) List(ctx context.Context) ([]*dbservice.DBPreset, error) {
	query := `
        SELECT id, type, size, mode, name, icon, description, friendly_description,
               technical_terms, use_cases, cpu, memory, disk, creation_cost, hourly_cost,
               default_config, sort_order, is_active
        FROM db_presets 
        WHERE is_active = true
        ORDER BY sort_order
    `

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	var presets []*dbservice.DBPreset
	for rows.Next() {
		var p dbservice.DBPreset
		var technicalTermsJSON, useCasesJSON, defaultConfigJSON []byte

		err := rows.Scan(
			&p.ID,
			&p.Type,
			&p.Size,
			&p.Mode,
			&p.Name,
			&p.Icon,
			&p.Description,
			&p.FriendlyDescription,
			&technicalTermsJSON, // JSONB → map[string]interface{}
			&useCasesJSON,       // JSONB → []string
			&p.Resources.CPU,
			&p.Resources.Memory,
			&p.Resources.Disk,
			&p.Cost.CreationCost,
			&p.Cost.HourlyLemons,
			&defaultConfigJSON, // JSONB → map[string]interface{}
			&p.SortOrder,
			&p.IsActive,
		)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		// JSON 파싱
		if len(technicalTermsJSON) > 0 {
			if err := json.Unmarshal(technicalTermsJSON, &p.TechnicalTerms); err != nil {
				return nil, errors.Wrap(err)
			}
		} else {
			p.TechnicalTerms = make(map[string]interface{})
		}

		if len(useCasesJSON) > 0 {
			if err := json.Unmarshal(useCasesJSON, &p.UseCases); err != nil {
				return nil, errors.Wrap(err)
			}
		} else {
			p.UseCases = []string{}
		}

		if len(defaultConfigJSON) > 0 {
			if err := json.Unmarshal(defaultConfigJSON, &p.DefaultConfig); err != nil {
				return nil, errors.Wrap(err)
			}
		} else {
			p.DefaultConfig = make(map[string]interface{})
		}

		presets = append(presets, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err)
	}

	return presets, nil
}
