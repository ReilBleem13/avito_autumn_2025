package service

type Service struct {
	users  UserRepositoryInterface
	teams  TeamRepositoryInterface
	prs    PullRequestRepositoryInterface
	logger LoggerInterfaces
}

func NewService(
	users UserRepositoryInterface,
	teams TeamRepositoryInterface,
	prs PullRequestRepositoryInterface,
	logger LoggerInterfaces,
) *Service {
	return &Service{
		users:  users,
		teams:  teams,
		prs:    prs,
		logger: logger,
	}
}
