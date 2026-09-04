package service

import (
	"fmt"
	"math/rand"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"
)

// QuizService 华裳问答 — 从服饰库自动出题
type QuizService struct {
	quizStore *model.QuizStore
}

// NewQuizService 创建问答 service
func NewQuizService(quizStore *model.QuizStore) *QuizService {
	return &QuizService{quizStore: quizStore}
}

// QuizQuestion 单道题目（不含答案字段，判分由服务端按 garment_id + field 重查）
type QuizQuestion struct {
	ID          string   `json:"id"`
	GarmentID   int64    `json:"garment_id"`
	Field       string   `json:"field"`
	Type        string   `json:"type"`
	Stem        string   `json:"stem"`
	Options     []string `json:"options"`
	Explanation string   `json:"explanation"`
}

// 字段静态候选池：数据池不足 3 个干扰项时补齐
var quizStaticPools = map[string][]string{
	"dynasty":  {"汉", "唐", "宋", "元", "明", "清", "现代"},
	"ethnic":   {"汉族", "满族", "蒙古族", "回族", "藏族", "维吾尔族", "苗族", "彝族", "壮族", "朝鲜族"},
	"province": {"北京", "陕西", "甘肃", "青海", "山东", "山西", "河南", "湖北", "广东", "福建", "江西", "安徽"},
}

// GenerateQuestions 生成 count 道题（5 种模板轮转，选项已洗牌且不含答案标记）
func (s *QuizService) GenerateQuestions(count int) ([]QuizQuestion, error) {
	sources, err := s.quizStore.ListSource()
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return []QuizQuestion{}, nil
	}

	// 各字段候选池（去重）
	pools := map[string][]string{
		"ethnic":   distinctField(sources, func(q model.QuizSource) string { return q.Ethnic }),
		"dynasty":  distinctField(sources, func(q model.QuizSource) string { return q.Dynasty }),
		"province": distinctField(sources, func(q model.QuizSource) string { return q.Province }),
	}
	names := make([]string, 0, len(sources))
	for _, q := range sources {
		names = append(names, q.Name)
	}

	shuffled := make([]model.QuizSource, len(sources))
	copy(shuffled, sources)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	questions := make([]QuizQuestion, 0, count)
	used := map[string]bool{} // 去重：garmentID:field
	idx := 0
	attempt := 0

	for len(questions) < count && attempt < count*8+len(sources) {
		attempt++
		q := shuffled[idx%len(shuffled)]
		idx++
		if idx > len(shuffled)*2 {
			break
		}

		templates := []string{"ethnic", "dynasty", "province", "name_craft", "name_pattern"}
		qType := templates[len(questions)%len(templates)]

		var field, answer, stem string
		var options []string

		switch qType {
		case "ethnic":
			if q.Ethnic == "" {
				continue
			}
			field, answer = "ethnic", q.Ethnic
			stem = fmt.Sprintf("「%s」来自哪个民族？", q.Name)
			options = buildOptions(answer, pools["ethnic"], quizStaticPools["ethnic"])
		case "dynasty":
			if q.Dynasty == "" {
				continue
			}
			field, answer = "dynasty", q.Dynasty
			stem = fmt.Sprintf("「%s」盛行于哪个朝代？", q.Name)
			options = buildOptions(answer, pools["dynasty"], quizStaticPools["dynasty"])
		case "province":
			if q.Province == "" {
				continue
			}
			field, answer = "province", q.Province
			stem = fmt.Sprintf("「%s」主要流传于哪个地区？", q.Name)
			options = buildOptions(answer, pools["province"], quizStaticPools["province"])
		case "name_craft":
			if q.Craft == "" {
				continue
			}
			field, answer = "name", q.Name
			stem = fmt.Sprintf("哪件服饰以「%s」工艺著称？", q.Craft)
			options = buildNameOptions(answer, names)
		case "name_pattern":
			if q.Pattern == "" {
				continue
			}
			field, answer = "name", q.Name
			stem = fmt.Sprintf("「%s」纹样见于哪件服饰？", q.Pattern)
			options = buildNameOptions(answer, names)
		}

		// 选项不足 4 个或题目重复则跳过
		if len(options) < 4 || used[quizKey(q.ID, field)] {
			continue
		}
		used[quizKey(q.ID, field)] = true

		questions = append(questions, QuizQuestion{
			ID:          fmt.Sprintf("q%d", len(questions)+1),
			GarmentID:   q.ID,
			Field:       field,
			Type:        qType,
			Stem:        stem,
			Options:     options,
			Explanation: buildExplanation(q),
		})
	}

	return questions, nil
}

// quizKey 去重键（garment + field）
func quizKey(garmentID int64, field string) string {
	return fmt.Sprintf("%d:%s", garmentID, field)
}

// distinctField 提取服饰列表中某字段的非空去重值
func distinctField(sources []model.QuizSource, get func(model.QuizSource) string) []string {
	seen := map[string]bool{}
	list := make([]string, 0)
	for _, q := range sources {
		v := strings.TrimSpace(get(q))
		if v != "" && !seen[v] {
			seen[v] = true
			list = append(list, v)
		}
	}
	return list
}

// buildOptions 从数据池 + 静态池中挑选干扰项，凑足 4 个选项并洗牌
func buildOptions(answer string, pools ...[]string) []string {
	seen := map[string]bool{answer: true}
	distractors := make([]string, 0, 3)
	for _, pool := range pools {
		shuffled := make([]string, len(pool))
		copy(shuffled, pool)
		rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		for _, v := range shuffled {
			if v == "" || seen[v] {
				continue
			}
			seen[v] = true
			distractors = append(distractors, v)
			if len(distractors) == 3 {
				break
			}
		}
		if len(distractors) == 3 {
			break
		}
	}

	options := append([]string{answer}, distractors...)
	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	return options
}

// buildNameOptions 服饰名选择题的干扰项（其余服饰名）
func buildNameOptions(answer string, names []string) []string {
	shuffled := make([]string, len(names))
	copy(shuffled, names)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	distractors := make([]string, 0, 3)
	for _, n := range shuffled {
		if n == answer || n == "" {
			continue
		}
		distractors = append(distractors, n)
		if len(distractors) == 3 {
			break
		}
	}
	options := append([]string{answer}, distractors...)
	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	return options
}

// buildExplanation 解析文案（答错也能涨知识）：文化寓意优先，简介兜底，截断 120 字
func buildExplanation(q model.QuizSource) string {
	text := strings.TrimSpace(q.CulturalMeaning)
	if text == "" {
		text = strings.TrimSpace(q.Description)
	}
	runes := []rune(text)
	if len(runes) > 120 {
		return string(runes[:120]) + "…"
	}
	return text
}
