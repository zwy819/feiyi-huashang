package ai

import "testing"

func TestParseVisualAttributes(t *testing.T) {
	attrs, err := parseVisualAttributes("```json\n{\"ethnic\":\"朝鲜族\",\"craft\":\"刺绣\",\"pattern\":\"花卉纹\",\"dynasty\":\"现代\",\"confidence\":0.9}\n```")
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Ethnic != "朝鲜族" || attrs.Craft != "刺绣" || attrs.Confidence != 0.9 {
		t.Fatalf("unexpected attributes: %+v", attrs)
	}
}

func TestVisualScorePrefersMatchingGarment(t *testing.T) {
	attrs := VisualAttributes{Ethnic: "朝鲜族", Craft: "刺绣", Pattern: "花卉纹", Dynasty: "现代"}
	item := VisualCandidate{GarmentID: 200, Name: "朝鲜族短袄裙", Ethnic: "朝鲜族", Craft: "刺绣", Dynasty: "现代"}
	score, reasons := visualScore(attrs, item, "丝绸+棉布", "花卉纹+蝴蝶纹", "朝鲜族女性服饰")
	if score < 8 {
		t.Fatalf("score = %f, want strong match; reasons=%v", score, reasons)
	}
}
