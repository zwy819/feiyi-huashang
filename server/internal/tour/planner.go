package tour

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Stop struct {
	Order           int    `json:"order"`
	GarmentID       int64  `json:"garment_id"`
	Name            string `json:"name"`
	Ethnic          string `json:"ethnic"`
	Craft           string `json:"craft"`
	Province        string `json:"province"`
	AnnotationID    int64  `json:"annotation_id"`
	AnnotationTitle string `json:"annotation_title"`
	DurationMin     int    `json:"duration_min"`
	Reason          string `json:"reason"`
}

type Plan struct {
	Title        string `json:"title"`
	Minutes      int    `json:"minutes"`
	TotalMinutes int    `json:"total_minutes"`
	Stops        []Stop `json:"stops"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

var minutesRE = regexp.MustCompile(`(\d{1,3})\s*分钟`)

var interestWords = []string{
	"刺绣", "银饰", "蜡染", "织锦", "挑花", "缂丝", "扎染", "印花",
	"苗", "藏", "彝", "傣", "侗", "壮", "满", "回", "土家", "朝鲜",
}

func (s *Service) Plan(ctx context.Context, query string) (Plan, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		query = "我有40分钟，想看最有代表性的非遗服饰"
	}
	minutes := 40
	if match := minutesRE.FindStringSubmatch(query); len(match) == 2 {
		if value, err := strconv.Atoi(match[1]); err == nil && value >= 10 && value <= 120 {
			minutes = value
		}
	}
	interests := make([]string, 0, 4)
	for _, word := range interestWords {
		if strings.Contains(query, word) {
			interests = append(interests, word)
		}
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, ethnic, craft, province, description, model_id FROM garments ORDER BY id`)
	if err != nil {
		return Plan{}, err
	}
	defer rows.Close()

	type candidate struct {
		id          int64
		name        string
		ethnic      string
		craft       string
		province    string
		description string
		modelID     sql.NullInt64
		score       int
	}
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.id, &c.name, &c.ethnic, &c.craft, &c.province, &c.description, &c.modelID); err != nil {
			return Plan{}, err
		}
		for _, interest := range interests {
			if strings.Contains(c.name+c.craft+c.description, interest) {
				c.score += 5
			}
		}
		if c.modelID.Valid {
			c.score += 1
		}
		if c.score == 0 && strings.Contains(query, "代表性") {
			c.score = 1
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return Plan{}, err
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].id < candidates[j].id
		}
		return candidates[i].score > candidates[j].score
	})

	selected := make([]candidate, 0, 5)
	usedEthnic := make(map[string]bool)
	for _, c := range candidates {
		if len(selected) >= 5 {
			break
		}
		if c.score > 0 && !usedEthnic[c.ethnic] {
			selected = append(selected, c)
			usedEthnic[c.ethnic] = true
		}
	}
	for _, c := range candidates {
		if len(selected) >= 5 {
			break
		}
		already := false
		for _, s := range selected {
			if s.id == c.id {
				already = true
				break
			}
		}
		if !already {
			selected = append(selected, c)
		}
	}

	perStop := minutes / 5
	if perStop < 4 {
		perStop = 4
	}
	stops := make([]Stop, 0, len(selected))
	for _, c := range selected {
		var annotationID int64
		var annotationTitle string
		err := s.db.QueryRowContext(ctx,
			`SELECT id, title FROM annotations WHERE garment_id = ? ORDER BY sort_order ASC, id ASC LIMIT 1`,
			c.id).Scan(&annotationID, &annotationTitle)
		if err != nil {
			continue
		}
		reason := "馆藏代表服饰"
		for _, interest := range interests {
			if strings.Contains(c.name+c.craft+c.description, interest) {
				reason = "匹配「" + interest + "」"
				break
			}
		}
		if c.modelID.Valid {
			reason += " · 可 3D 观看"
		}
		stops = append(stops, Stop{
			Order:           len(stops) + 1,
			GarmentID:       c.id,
			Name:            c.name,
			Ethnic:          c.ethnic,
			Craft:           c.craft,
			Province:        c.province,
			AnnotationID:    annotationID,
			AnnotationTitle: annotationTitle,
			DurationMin:     perStop,
			Reason:          reason,
		})
	}
	if len(stops) == 0 {
		return Plan{}, fmt.Errorf("no garment annotations available")
	}

	title := fmt.Sprintf("%d 分钟 · 非遗服饰参观路线", minutes)
	if len(interests) > 0 {
		title = fmt.Sprintf("%d 分钟 · %s路线", minutes, strings.Join(interests, " + "))
	}
	return Plan{
		Title:        title,
		Minutes:      minutes,
		TotalMinutes: perStop * len(stops),
		Stops:        stops,
	}, nil
}
