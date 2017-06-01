package service
import "errors"

type AppConfig struct {
	AWSRegion string
	DynamoDbTableName string
	appSystemCode string
	appName string
	port string
}

type Model struct {
	UUID string `json:"conceptId"`
	ConcordedIds []string `json:"concordedIds`
}

type Service interface {
	Read(uuid string) (Model, error)
	Write(m Model) (bool, error)
	Delete(uuid string) (bool, error)
	Count() (int, error)
}

type ConcordancesRwService struct {
	conf AppConfig
}



func NewConcordancesRwService(conf AppConfig) Service {
	s := ConcordancesRwService{conf: conf}

	return &s;

}

func (s *ConcordancesRwService) Read(uuid string) (Model, error){

	err := errors.New("Not Implemented")
	return Model{}, err
}

func (h *ConcordancesRwService) Write(m Model) (bool, error) {

	err := errors.New("Not Implemented")
	return false, err
}

func (h *ConcordancesRwService) Delete(uuid string) (bool, error) {

	err := errors.New("Not Implemented")
	return false, err
}

func (h *ConcordancesRwService) Count() (int, error) {

	err := errors.New("Not Implemented")
	return 0, err
}

