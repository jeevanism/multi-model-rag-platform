package service

import "testing"

type fakeEvalRepo struct {
	runs   []EvalRunSummary
	detail EvalRunDetailResult
	found  bool
}

func (f fakeEvalRepo) ListEvalRuns(limit int) ([]EvalRunSummary, error) {
	return f.runs, nil
}

func (f fakeEvalRepo) GetEvalRunDetail(evalRunID int) (EvalRunDetailResult, bool, error) {
	return f.detail, f.found, nil
}

func TestListEvalRunsReturnsData(t *testing.T) {
	service := NewEvalService(fakeEvalRepo{
		runs: []EvalRunSummary{
			{ID: 1, DatasetName: "eval_set.jsonl", Provider: "gemini", TotalCases: 3, PassedCases: 3, CreatedAt: "2026-02-23T00:00:00+00:00"},
		},
	})

	runs, err := service.ListEvalRuns(20)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(runs) != 1 || runs[0].ID != 1 {
		t.Fatalf("unexpected runs %#v", runs)
	}
}

func TestGetEvalRunDetailReturnsNotFoundError(t *testing.T) {
	service := NewEvalService(fakeEvalRepo{found: false})

	_, err := service.GetEvalRunDetail(999)
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrEvalRunNotFound {
		t.Fatalf("expected ErrEvalRunNotFound, got %v", err)
	}
}
