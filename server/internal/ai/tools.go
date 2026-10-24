package ai

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"
)

type FilterSet struct {
	Ethnic   string
	Province string
	Craft    string
	Dynasty  string
}

func (f FilterSet) Empty() bool {
	return f.Ethnic == "" && f.Province == "" && f.Craft == "" && f.Dynasty == ""
}

func (f FilterSet) Summary() string {
	var parts []string
	if f.Ethnic != "" {
		parts = append(parts, "民族="+f.Ethnic)
	}
	if f.Province != "" {
		parts = append(parts, "省份="+f.Province)
	}
	if f.Craft != "" {
		parts = append(parts, "工艺="+f.Craft)
	}
	if f.Dynasty != "" {
		parts = append(parts, "朝代="+f.Dynasty)
	}
	return strings.Join(parts, "，")
}

type GarmentBrief struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Dynasty  string `json:"dynasty"`
	Ethnic   string `json:"ethnic"`
	Craft    string `json:"craft"`
	Province string `json:"province"`
	ModelID  *int64 `json:"model_id,omitempty"`
}

type StatsItem struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type Toolbox struct {
	db              *sql.DB
	garmentStore    *model.GarmentStore
	annotationStore *model.AnnotationStore
}

func NewToolbox(db *sql.DB, garmentStore *model.GarmentStore, annotationStore *model.AnnotationStore) *Toolbox {
	return &Toolbox{db: db, garmentStore: garmentStore, annotationStore: annotationStore}
}

// ExtractFilters matches explicit ethnic/province/craft/dynasty words in the
// question against values that actually exist in the collection. This keeps
// SQL filtering grounded and avoids inventing categories the model might know.
func (t *Toolbox) ExtractFilters(ctx context.Context, query string) (FilterSet, error) {
	filters := FilterSet{}
	dimensions := []struct {
		column string
		target *string
	}{
		{"ethnic", &filters.Ethnic},
		{"province", &filters.Province},
		{"craft", &filters.Craft},
		{"dynasty", &filters.Dynasty},
	}
	for _, dim := range dimensions {
		rows, err := t.db.QueryContext(ctx, fmt.Sprintf(`SELECT DISTINCT %s FROM garments WHERE %s != ''`, dim.column, dim.column))
		if err != nil {
			return filters, err
		}
		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				rows.Close()
				return filters, err
			}
			// Craft cells sometimes contain "蜡染+刺绣"; match each part.
			parts := []string{value}
			if dim.column == "craft" {
				parts = strings.FieldsFunc(value, func(r rune) bool {
					return r == '+' || r == '、' || r == '/' || r == ' '
				})
			}
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" && strings.Contains(query, part) && len([]rune(part)) > len([]rune(*dim.target)) {
					*dim.target = part
				}
			}
		}
		rows.Close()
	}
	return filters, nil
}

func (t *Toolbox) FindGarmentByName(ctx context.Context, query string) (*GarmentBrief, error) {
	rows, err := t.db.QueryContext(ctx, `SELECT id, name, dynasty, ethnic, craft, province, model_id FROM garments ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var best *GarmentBrief
	for rows.Next() {
		var g GarmentBrief
		if err := rows.Scan(&g.ID, &g.Name, &g.Dynasty, &g.Ethnic, &g.Craft, &g.Province, &g.ModelID); err != nil {
			return nil, err
		}
		if strings.Contains(query, g.Name) {
			copy := g
			if best == nil || len([]rune(copy.Name)) > len([]rune(best.Name)) {
				best = &copy
			}
		}
	}
	return best, rows.Err()
}

func (t *Toolbox) SearchGarments(ctx context.Context, filters FilterSet, name string, limit int) ([]GarmentBrief, error) {
	if limit <= 0 || limit > 12 {
		limit = 8
	}
	where := []string{"1=1"}
	args := []any{}
	if filters.Ethnic != "" {
		where = append(where, "ethnic = ?")
		args = append(args, filters.Ethnic)
	}
	if filters.Province != "" {
		where = append(where, "province = ?")
		args = append(args, filters.Province)
	}
	if filters.Craft != "" {
		where = append(where, "craft LIKE ?")
		args = append(args, "%"+filters.Craft+"%")
	}
	if filters.Dynasty != "" {
		where = append(where, "dynasty = ?")
		args = append(args, filters.Dynasty)
	}
	if name != "" {
		where = append(where, "(name LIKE ? OR description LIKE ? OR pattern LIKE ?)")
		like := "%" + name + "%"
		args = append(args, like, like, like)
	}
	args = append(args, limit)
	query := `SELECT id, name, dynasty, ethnic, craft, province, model_id FROM garments
		WHERE ` + strings.Join(where, " AND ") + ` ORDER BY id DESC LIMIT ?`
	rows, err := t.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]GarmentBrief, 0, limit)
	for rows.Next() {
		var g GarmentBrief
		if err := rows.Scan(&g.ID, &g.Name, &g.Dynasty, &g.Ethnic, &g.Craft, &g.Province, &g.ModelID); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (t *Toolbox) GetGarment(ctx context.Context, id int64) (*model.Garment, error) {
	return t.garmentStore.GetByID(id)
}

func (t *Toolbox) GetAnnotations(ctx context.Context, garmentID int64) ([]model.Annotation, error) {
	return t.annotationStore.ListByGarment(garmentID)
}

func (t *Toolbox) GetStatistics(ctx context.Context, dimension string) ([]StatsItem, error) {
	column := ""
	switch dimension {
	case "province", "ethnic", "craft", "dynasty":
		column = dimension
	default:
		return nil, fmt.Errorf("unsupported statistics dimension %q", dimension)
	}
	rows, err := t.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT %s, COUNT(*) AS n FROM garments WHERE %s != ''
		GROUP BY %s ORDER BY n DESC, %s ASC LIMIT 8`, column, column, column, column))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]StatsItem, 0, 8)
	for rows.Next() {
		var item StatsItem
		if err := rows.Scan(&item.Label, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (t *Toolbox) Similar(ctx context.Context, garmentID int64) ([]GarmentBrief, error) {
	g, err := t.garmentStore.GetByID(garmentID)
	if err != nil {
		return nil, err
	}
	items, err := t.garmentStore.GetByEthnicAndCraft(g.Ethnic, g.Craft, garmentID)
	if err != nil {
		return nil, err
	}
	out := make([]GarmentBrief, 0, len(items))
	for _, item := range items {
		out = append(out, GarmentBrief{
			ID: item.ID, Name: item.Name, Dynasty: item.Dynasty,
			Ethnic: item.Ethnic, Craft: item.Craft, Province: item.Province, ModelID: item.ModelID,
		})
	}
	return out, nil
}
