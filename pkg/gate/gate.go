package gate

import (
	"APIGateway/pkg/obj"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
)

const (
	newsAggregator  = "http://localhost:8080"
	commentsService = "http://localhost:9595"
)

// делает запрос в сервис новостей и возвращает массив новостей
func GetLatestNews(ctx context.Context, p int) (any, error) {
	r, err := http.Get(newsAggregator + "/news/" + strconv.Itoa(p*15) + "?requestID=" + getRequestID(ctx))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Раскодируем JSON в массив новостей+requestID.
	var data = struct {
		Posts     []obj.NewsShortDetailed
		RequestID any
	}{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func PostComment(ctx context.Context, c obj.Comment) (any, error) {

	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(b)

	r, err := http.Post(commentsService+"/add"+"?requestID="+getRequestID(ctx), "application/json", buf)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	// Проверяем код ответа.
	if !(r.StatusCode == http.StatusOK) {
		return nil, fmt.Errorf("код ответ сервиса комментариев при попытке создать новый: %d", r.StatusCode)
	}
	var data = struct{
		RequestID any
	}{}

	// Читаем тело ответа.
	b, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetComments(ctx context.Context, id int) ([]obj.Comment, error) {
	r, err := http.Get(commentsService + "/comments?postID=" + strconv.Itoa(id)+"&requestID="+getRequestID(ctx))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Раскодируем JSON в массив комментариев+id запроса
	var data = struct {
		Comments      []obj.Comment
		RequestID any
	}{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data.Comments, nil
}

func GetPost(ctx context.Context, id int) (*obj.NewsFullDetailed, error) {
	r, err := http.Get(newsAggregator + "/news?postID=" + strconv.Itoa(id)+"&requestID="+getRequestID(ctx))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Раскодируем JSON 
	var data = struct {
		Post      obj.NewsFullDetailed
		RequestID any
	}{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &data.Post, nil
}

// отправляет 2 асинхронных запроса - в сервис новостей и сервис комментариев и готовит объект подробной новости
func GetDetailedPost(ctx context.Context, id int) (any, error) {
	c := make(chan interface{}, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		var r commentsResponse
		r.comments, r.err = GetComments(ctx,id)
		c <- r
	}()

	go func() {
		defer wg.Done()
		var r postResponse
		r.post, r.err = GetPost(ctx,id)
		c <- r
	}()

	wg.Wait()
	close(c)

	var r obj.NewsFullDetailed
	var com []obj.Comment

	for m := range c {
		switch m.(type) {
		case commentsResponse:
			a := m.(commentsResponse)
			if a.err != nil {
				return nil, a.err
			}
			com = a.comments
		case postResponse:
			a := m.(postResponse)
			if a.err != nil {
				return nil, a.err
			}
			r = *a.post
		}
	}
	r.Comment = com

	var ans = struct {
		obj.NewsFullDetailed
		RequestID any
	}{
		NewsFullDetailed: r,
		RequestID:       ctx.Value(obj.ContextKey("requestID")),
	}

	return &ans, nil
}

type commentsResponse struct {
	comments []obj.Comment
	err      error
}

type postResponse struct {
	post *obj.NewsFullDetailed
	err  error
}

// запрашивает сервис аггрегатора новостей с поисковым запросом
func SearchPosts(ctx context.Context, searchParam string, pageParam string) (any, error) {

	reqStr := newsAggregator + "/news?requestID=" + getRequestID(ctx)

	if searchParam != "" {
		reqStr += "&search=" + searchParam
	}

	if pageParam != "" {
		reqStr += "&page=" + pageParam
	}

	r, err := http.Get(reqStr)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Читаем тело ответа.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Раскодируем JSON в массив новостей + объект пагинации+ id запроса.

	var data = struct {
		Posts      []obj.NewsShortDetailed
		Pagination obj.Pagination
		RequestID  any
	}{}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func getRequestID(ctx context.Context) string {
	return strconv.Itoa(ctx.Value(obj.ContextKey("requestID")).(int))
}
