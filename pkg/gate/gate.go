package gate

import (
	"APIGateway/pkg/obj"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	newsAggregator = "http://localhost:8080/news"
	commentsService = "http://localhost:9595"
)

// делает запрос в сервис новостей и возвращает массив новостей
func GetLatestNews(p int) ([]obj.NewsShortDetailed, error) {
	r, err := http.Get(newsAggregator + "/" + strconv.Itoa(p*10))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Раскодируем JSON в массив новостей.
	var data []obj.NewsShortDetailed
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func PostComment(c obj.Comment) error {
	
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(b)

	r, err := http.Post(commentsService + "/add", "application/json", buf)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	// Проверяем код ответа.
	if !(r.StatusCode == http.StatusOK) {
		return fmt.Errorf("код ответ сервиса комментариев при попытке создать новый: %d", r.StatusCode)
	}
	return nil
}