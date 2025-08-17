package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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

func (s *PresetStore) Find(ctx context.Context, id string) (*dbservice.DBPreset, error) {
	const query = `
        SELECT 
            id, type, size, mode, name, icon, description, friendly_description,
            technical_terms, use_cases, cpu, memory, disk, creation_cost, hourly_cost,
            default_config, sort_order, available, unavailable_reason
        FROM db_presets
        WHERE id = $1
    `

	row := s.db.QueryRowContext(ctx, query, id)
	preset, err := s.scanPreset(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewResourceNotFoundError("preset", id)
		}
		return nil, fmt.Errorf("find preset: %w", err)
	}

	return preset, nil
}

func (s *PresetStore) ListByType(ctx context.Context, dbType dbservice.DBType) ([]*dbservice.DBPreset, error) {
	const query = `
        SELECT 
            id, type, size, mode, name, icon, description, friendly_description,
            technical_terms, use_cases, cpu, memory, disk, creation_cost, hourly_cost,
            default_config, sort_order, available, unavailable_reason
        FROM db_presets 
        WHERE type = $1
        ORDER BY sort_order, id
    `

	rows, err := s.db.QueryContext(ctx, query, dbType)
	if err != nil {
		return nil, fmt.Errorf("list presets by type: %w", err)
	}
	defer rows.Close()

	presets := make([]*dbservice.DBPreset, 0, 5)

	for rows.Next() {
		preset, err := s.scanPreset(rows)
		if err != nil {
			return nil, fmt.Errorf("scan preset: %w", err)
		}
		presets = append(presets, preset)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return presets, nil
}

func (s *PresetStore) scanPreset(scanner interface{ Scan(...interface{}) error }) (*dbservice.DBPreset, error) {
	var (
		preset             dbservice.DBPreset
		technicalTermsJSON []byte
		defaultConfigJSON  []byte
		useCasesArr        pq.StringArray
		cpu                float64
		unavailableReason  sql.NullString
	)

	err := scanner.Scan(
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
		&cpu,
		&preset.Resources.Memory,
		&preset.Resources.Disk,
		&preset.Cost.CreationCost,
		&preset.Cost.HourlyLemons,
		&defaultConfigJSON,
		&preset.SortOrder,
		&preset.Available,
		&unavailableReason,
	)
	if err != nil {
		return nil, err
	}

	preset.Resources.CPU = cpu
	preset.UnavailableReason = unavailableReason.String

	// JSONB 필드 파싱
	if err := s.parseJSONFields(&preset, technicalTermsJSON, defaultConfigJSON); err != nil {
		return nil, err
	}

	preset.UseCases = []string(useCasesArr)

	return &preset, nil
}

func (s *PresetStore) parseJSONFields(preset *dbservice.DBPreset, technicalTermsJSON, defaultConfigJSON []byte) error {
	// technical_terms 파싱
	if len(technicalTermsJSON) > 0 {
		if err := json.Unmarshal(technicalTermsJSON, &preset.TechnicalTerms); err != nil {
			return fmt.Errorf("unmarshal technical terms: %w", err)
		}
	} else {
		preset.TechnicalTerms = make(map[string]interface{})
	}

	// default_config 파싱
	if len(defaultConfigJSON) > 0 {
		if err := json.Unmarshal(defaultConfigJSON, &preset.DefaultConfig); err != nil {
			return fmt.Errorf("unmarshal default config: %w", err)
		}
	} else {
		preset.DefaultConfig = make(map[string]interface{})
	}

	return nil
}
