package service

import "cbr-worker/internal/cbr"

type Service struct {
	repo *cbr.Repository
}

func New(repo *cbr.Repository) *Service {
	return &Service{repo: repo}
}
