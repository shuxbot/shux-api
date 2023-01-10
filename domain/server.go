package domain

type ServerRepository interface {
	List()([]string, error)
	GetRanking(ServerId string) ([]User, error)
}