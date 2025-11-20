package repository

import (
	"ReilBleem13/pull_requests_service/internal/repository/database"

	"github.com/jmoiron/sqlx"
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *database.PostgresDB) *PullRequestRepository {
	return &PullRequestRepository{
		db: db.Client(),
	}
}

func (p *PullRequestRepository) Create() {}

func (p *PullRequestRepository) Merge() {}

func (p *PullRequestRepository) ReAssign() {}
