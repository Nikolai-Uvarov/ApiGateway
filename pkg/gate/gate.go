package gate

import (
	"APIGateway/pkg/obj"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
)

const (
	newsAggregator = "http://localhost:8080"
	commentsService = "http://localhost:9595"
)

// делает запрос в сервис новостей и возвращает массив новостей
func GetLatestNews(p int) ([]obj.NewsShortDetailed, error) {
	r, err := http.Get(newsAggregator + "/news/" + strconv.Itoa(p*10))
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

func GetComments(id int) ([]obj.Comment, error){
	r, err := http.Get(commentsService + "/comments?postID=" + strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Раскодируем JSON в массив комментариев.
	var data []obj.Comment
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetPost(id int) (*obj.NewsFullDetailed, error){
	r, err := http.Get(newsAggregator + "/news?postID=" + strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Раскодируем JSON в массив комментариев.
	var data obj.NewsFullDetailed
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

//отправляет 2 асинхронных запроса - в сервис новостей и сервис комментариев и готовит объект подробной новости
func GetDetailedPost(id int) (*obj.NewsFullDetailed, error){
	c:=make(chan interface{},2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func(){
		defer wg.Done()
		var r commentsResponse
		r.comments, r.err = GetComments(id)
		c<-r
	}()

	go func(){
		defer wg.Done()
		var r postResponse
		r.post, r.err = GetPost(id)
		c<-r
	}()

	wg.Wait()
	close(c)

	var r obj.NewsFullDetailed
	var com []obj.Comment

	for m:=range c {
		switch m.(type){
		case commentsResponse:
			a:=m.(commentsResponse)
			if a.err!=nil{
				return nil, a.err
			}
			com = a.comments
		case postResponse:
			a:=m.(postResponse)		
			if a.err!=nil{
				return nil, a.err
			}	
			r = *a.post
		}
	}
	r.Comment=com
	return &r,nil
}

type commentsResponse struct{
	comments []obj.Comment
	err error
}

type postResponse struct{
	post *obj.NewsFullDetailed
	err error
}