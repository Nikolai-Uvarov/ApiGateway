package api

import (
	"APIGateway/pkg/obj"
	"APIGateway/pkg/gate"
	"encoding/json"
	"log"

	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// API сервиса.
type API struct {
	r *mux.Router // маршрутизатор запросов
}

// Конструктор API.
func New() *API {
	api := API{}
	api.r = mux.NewRouter()
	api.endpoints()
	return &api
}

// Router возвращает маршрутизатор запросов.
func (api *API) Router() *mux.Router {
	return api.r
}

// Регистрация методов API в маршрутизаторе запросов.
func (api *API) endpoints() {
	// получить n последних новостей
	api.r.HandleFunc("/news/latest", api.posts).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/filter", api.filter).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/post", api.postByID).Methods(http.MethodGet, http.MethodOptions)
	api.r.HandleFunc("/news/comment", api.addComment).Methods(http.MethodPost, http.MethodOptions)
	//заголовок ответа
	api.r.Use(api.HeadersMiddleware)
}

// HeadersMiddleware устанавливает заголовки ответа сервера.
func (api *API) HeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// posts возвращает n-ую страницу (страница = 10 постов) новейших новостей в зависимости от параметра page=n
func (api *API) posts(w http.ResponseWriter, r *http.Request) {
	// Считывание параметра page строки запроса.

	// если параметр был передан, вернется строка со значением.
	// Если не был - в переменной будет пустая строка
	pageParam := r.URL.Query().Get("page")
	// параметр page - это число, поэтому нужно сконвертировать
	// строку в число при помощи пакета strconv
	var page int
	var err error

	if pageParam != "" {
		page, err = strconv.Atoi(pageParam)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		page = 1
	}
	// Получение данных из сервиса новостей
	o, err:=gate.GetLatestNews(page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Отправка данных клиенту в формате JSON.
	json.NewEncoder(w).Encode(o)
	// Отправка клиенту статуса успешного выполнения запроса
	w.WriteHeader(http.StatusOK)
}

var FullNews = []obj.NewsFullDetailed{
	{ID: 1,
		Title:   "Важная новость",
		Content: "Очень важные события произошли с очень важными людьми при очень необычных обстоятельствах",
		PubTime: 1696349769,
		Link:    "www.rbc.ru"},
	{ID: 2,
		Title:   "Не менее важная новость",
		Content: "Не менее важные события произошли с не менее важными людьми при не менее необычных обстоятельствах",
		PubTime: 1696349770,
		Link:    "www.kommersant.ru",
		Comment: obj.Comment{}},
}

var ShortNews = []obj.NewsShortDetailed{
	{ID: 1,
		Title:   "Важная новость",
		PubTime: 1696349769,
		Link:    "www.rbc.ru"},
	{ID: 2,
		Title:   "Не менее важная новость",
		PubTime: 1696349770,
		Link:    "www.kommersant.ru"},
}

// Фильтр или поиск новостей.
// Для данного метода параметры:
// contains = слово -  совпадение слова в заголовке новости
// dateafter, datebefore = UNIX time -  выбор диапазона дат,
// notcontains = слово -  слова в заголовке, которые исключить,
// sort = date/name - выбор поля для сортировки (по дате, по названию).
func (api *API) filter(w http.ResponseWriter, r *http.Request) {
	// Считывание параметров фильтра page строки запроса.
	containsParam := r.URL.Query().Get("contains")
	notcontainsParam := r.URL.Query().Get("notcontains")
	dafterParam := r.URL.Query().Get("dateafter")
	dbeforeParam := r.URL.Query().Get("datebefore")
	sortParam := r.URL.Query().Get("sort")

	// Получение данных из сервиса новостей или из кеша - пока замокано
	posts := ShortNews
	log.Println(containsParam, notcontainsParam, dafterParam, dbeforeParam, sortParam)
	// Отправка данных клиенту в формате JSON.
	json.NewEncoder(w).Encode(posts)
	// Отправка клиенту статуса успешного выполнения запроса
	w.WriteHeader(http.StatusOK)
}

// возвращает детальную новость по postID
func (api *API) postByID(w http.ResponseWriter, r *http.Request) {
	// Считывание параметра  строки запроса.
	idParam := r.URL.Query().Get("postID")

	id, err := strconv.Atoi(idParam)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Получение данных из сервиса новостей или из кеша - пока замокано
	post := FullNews[0]
	log.Println(id)
	// Отправка данных клиенту в формате JSON.
	json.NewEncoder(w).Encode(post)
	// Отправка клиенту статуса успешного выполнения запроса
	w.WriteHeader(http.StatusOK)
}

// метод добавления комментария
// Принимает ID новости postID или ID родительского комментария commentID и текст комментария в теле запроса
func (api *API) addComment(w http.ResponseWriter, r *http.Request) {

	var c obj.Comment
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//тут отправка запроса на создание комментария  в сервис комментариев
	err = gate.PostComment(c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Отправка клиенту статуса успешного выполнения запроса
	w.WriteHeader(http.StatusOK)
}
