package repository

type UserRepository interface {
	GetEmailByID(userID string) (string, error)
	SaveEmail(userID, email string) error
}

type InMemoryUserRepo struct {
	data map[string]string
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		data: make(map[string]string),
	}
}

func (r *InMemoryUserRepo) GetEmailByID(userID string) (string, error) {
	if e, ok := r.data[userID]; ok {
		return e, nil
	}
	return "", nil
}

func (r *InMemoryUserRepo) SaveEmail(userID, email string) error {
	r.data[userID] = email
	return nil
}
