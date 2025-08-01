package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
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

func (s *PresetStore) Detail(ctx context.Context, id string) (*dbservice.DBPreset, error) {
	const query = `
        SELECT 
            id, type, size, mode, name, icon, description, friendly_description,
            technical_terms, use_cases, cpu, memory, disk, creation_cost, hourly_cost,
            default_config, sort_order
        FROM db_presets
        WHERE id = $1 AND is_active = true
    `

	preset := &dbservice.DBPreset{}
	var technicalTermsJSON, defaultConfigJSON []byte
	var useCasesArr pq.StringArray

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&preset.ID,
		&preset.Type,
		&preset.Size,
		&preset.Mode,
		&preset.Name,
		&preset.Icon,
		&preset.Description,
		&preset.FriendlyDescription,
		&technicalTermsJSON,
		&useCasesArr,
		&preset.Resources.CPU,
		&preset.Resources.Memory,
		&preset.Resources.Disk,
		&preset.Cost.CreationCost,
		&preset.Cost.HourlyLemons,
		&defaultConfigJSON,
		&preset.SortOrder,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewResourceNotFoundError("preset", id)
		}
		return nil, errors.Wrap(err)
	}

	if err := s.parseJSONFields(preset, technicalTermsJSON, defaultConfigJSON); err != nil {
		return nil, err
	}

	// PostgreSQL array -> Go slice
	preset.UseCases = []string(useCasesArr)

	return preset, nil
}

func (s *PresetStore) ListByType(ctx context.Context, dbType dbservice.DBType) ([]*dbservice.DBPreset, error) {
	const query = `
        SELECT 
            id, type, size, mode, name, icon, description, friendly_description,
            technical_terms, use_cases, cpu, memory, disk, creation_cost, hourly_cost,
            default_config, sort_order
        FROM db_presets 
        WHERE type = $1 AND is_active = true
        ORDER BY sort_order, id
    `

	rows, err := s.db.QueryContext(ctx, query, dbType)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	presets := make([]*dbservice.DBPreset, 0)

	for rows.Next() {
		preset, err := s.scanPreset(rows)
		if err != nil {
			return nil, err
		}
		presets = append(presets, preset)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err)
	}

	return presets, nil
}

func (s *PresetStore) scanPreset(rows *sql.Rows) (*dbservice.DBPreset, error) {
	var p dbservice.DBPreset
	var technicalTermsJSON, defaultConfigJSON []byte
	var useCasesArr pq.StringArray

	err := rows.Scan(
		&p.ID,
		&p.Type,
		&p.Size,
		&p.Mode,
		&p.Name,
		&p.Icon,
		&p.Description,
		&p.FriendlyDescription,
		&technicalTermsJSON,
		&useCasesArr,
		&p.Resources.CPU,
		&p.Resources.Memory,
		&p.Resources.Disk,
		&p.Cost.CreationCost,
		&p.Cost.HourlyLemons,
		&defaultConfigJSON,
		&p.SortOrder,
	)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// Parse JSONB fields
	if err := s.parseJSONFields(&p, technicalTermsJSON, defaultConfigJSON); err != nil {
		return nil, err
	}

	// Convert PostgreSQL array to Go slice
	p.UseCases = []string(useCasesArr)

	return &p, nil
}

func (s *PresetStore) parseJSONFields(preset *dbservice.DBPreset, technicalTermsJSON, defaultConfigJSON []byte) error {
	// Parse technical_terms
	if len(technicalTermsJSON) > 0 {
		if err := json.Unmarshal(technicalTermsJSON, &preset.TechnicalTerms); err != nil {
			return errors.Wrap(err)
		}
	} else {
		preset.TechnicalTerms = make(map[string]interface{})
	}

	// Parse default_config
	if len(defaultConfigJSON) > 0 {
		if err := json.Unmarshal(defaultConfigJSON, &preset.DefaultConfig); err != nil {
			return errors.Wrap(err)
		}
	} else {
		preset.DefaultConfig = make(map[string]interface{})
	}

	return nil
}
