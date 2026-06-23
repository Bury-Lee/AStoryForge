package test_function

import (
	"AStoryForge/function/story_struct"
	"encoding/json"
	"os"
	"testing"
)

func TestVerifySunflower2(t *testing.T) {
	data, _ := os.ReadFile("向日葵2.json")
	var projObj story_struct.ProjectObj
	if err := json.Unmarshal(data, &projObj); err != nil {
		t.Fatal(err)
	}
	for evID, ev := range projObj.Events {
		if ev.ID == "" || ev.Name == "" || ev.Process == "" || len(ev.Outcome) == 0 {
			t.Errorf("event %s incomplete", evID)
		}
	}
	for entID, ent := range projObj.Entities {
		if ent.ID == "" || ent.Name == "" || len(ent.Introduction) == 0 {
			t.Errorf("entity %s incomplete", entID)
		}
	}
	for _, evID := range projObj.WorldSetting.Spine {
		if _, ok := projObj.Events[evID]; !ok {
			t.Errorf("spine ref %s missing", evID)
		}
	}
}
