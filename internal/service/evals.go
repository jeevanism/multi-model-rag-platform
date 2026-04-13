package service

import (
	"errors"
)

var ErrEvalRunNotFound = errors.New("Eval run not found")

type EvalRepository interface {
	ListEvalRuns(limit int) ([]EvalRunSummary, error)
	GetEvalRunDetail(evalRunID int) (EvalRunDetailResult, bool, error)
}

type EvalRunSummary struct {
	ID           int
	DatasetName  string
	Provider     string
	Model        *string
	TotalCases   int
	PassedCases  int
	AvgLatencyMS *float64
	CreatedAt    string
}

type EvalRunCase struct {
	ID                 int
	CaseID             string
	Question           string
	Passed             bool
	LatencyMS          int
	CorrectnessScore   *float64
	GroundednessScore  *float64
	HallucinationScore *float64
	Citations          []string
	Error              *string
}

type EvalRunDetailResult struct {
	Run   EvalRunSummary
	Cases []EvalRunCase
}

type EvalService struct {
	repo EvalRepository
}

func NewEvalService(repo EvalRepository) EvalService {
	return EvalService{repo: repo}
}

func (s EvalService) ListEvalRuns(limit int) ([]EvalRunSummary, error) {
	return s.repo.ListEvalRuns(limit)
}

func (s EvalService) GetEvalRunDetail(evalRunID int) (EvalRunDetailResult, error) {
	detail, found, err := s.repo.GetEvalRunDetail(evalRunID)
	if err != nil {
		return EvalRunDetailResult{}, err
	}
	if !found {
		return EvalRunDetailResult{}, ErrEvalRunNotFound
	}
	return detail, nil
}
