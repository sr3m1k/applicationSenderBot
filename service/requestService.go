package service

import "applicationBot/repoRequests"

type RequestService struct {
	requestRepo *repoRequests.RequestRepository
}

func NewRequestService(requestRepo *repoRequests.RequestRepository) *RequestService {
	return &RequestService{requestRepo: requestRepo}
}

func (s *RequestService) CreateRequest(number int, comment string, userId int64, username string, datetime string) error {
	request := repoRequests.Request{
		Number:   number,
		Comment:  comment,
		UserId:   int(userId),
		Username: username,
		Datetime: datetime,
	}
	return s.requestRepo.AddRequest(request)
}

func (s *RequestService) GetRequestsByUserId(userId int64) ([]repoRequests.Request, error) {
	return s.requestRepo.GetRequestsByUserId(userId)
}

func (s *RequestService) GetRequestDetails(number int) (repoRequests.Request, error) {
	return s.requestRepo.GetRequestByNumber(number)
}

func (s *RequestService) DeleteRequest(number int) error {
	return s.requestRepo.DeleteRequestByNumber(number)
}

func (s *RequestService) AddChat(table string, chatID int64, chatTitle string) error {
	return s.requestRepo.AddChatToDB(table, chatID, chatTitle)
}
