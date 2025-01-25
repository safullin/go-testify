package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cafeList = map[string][]string{
	"moscow": []string{"Мир кофе", "Сладкоежка", "Кофе и завтраки", "Сытый студент"},
}

func mainHandle(w http.ResponseWriter, req *http.Request) {
	countStr := req.URL.Query().Get("count")
	if countStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("count missing"))
		return
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong count value"))
		return
	}

	city := req.URL.Query().Get("city")

	cafe, ok := cafeList[city]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong city value"))
		return
	}

	if count > len(cafe) {
		count = len(cafe)
	}

	answer := strings.Join(cafe[:count], ",")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(answer))
}

func TestMainHandlerSuccess(t *testing.T) {
	// 1. Создаём GET-запрос к /cafe с city=moscow и count=1
	req, err := http.NewRequest("GET", "/cafe?count=1&city=moscow", nil)
	require.NoError(t, err)

	// 2. Создаём recorder (запись) для фиктивного HTTP-ответа
	rr := httptest.NewRecorder()

	// 3. Создаём обработчик и отправляем запрос
	handler := http.HandlerFunc(mainHandle)
	handler.ServeHTTP(rr, req)

	// 4. Проверки
	//    Код ответа должен быть 200
	require.Equal(t, http.StatusOK, rr.Code)
	//    Тело ответа не должно быть пустым
	assert.NotEmpty(t, rr.Body.String())
}

func TestMainHandlerWrongCity(t *testing.T) {
	// 1. Создаём GET-запрос к /cafe с city=unknown и count=1
	req, err := http.NewRequest("GET", "/cafe?count=1&city=unknown", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mainHandle)
	handler.ServeHTTP(rr, req)

	// Код ответа должен быть 400
	require.Equal(t, http.StatusBadRequest, rr.Code)
	// Тело ответа должно содержать ошибку "wrong city value"
	assert.Equal(t, "wrong city value", rr.Body.String())
}

func TestMainHandlerWhenCountMoreThanTotal(t *testing.T) {
	// Предположим, в "moscow" всего 4 кафе,
	// а мы запрашиваем 10
	req, err := http.NewRequest("GET", "/cafe?count=10&city=moscow", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mainHandle)
	handler.ServeHTTP(rr, req)

	// Код ответа должен быть 200
	require.Equal(t, http.StatusOK, rr.Code)

	// Ожидается, что вернутся все 4 кафе, соединённые запятой
	expected := "Мир кофе,Сладкоежка,Кофе и завтраки,Сытый студент"
	assert.Equal(t, expected, rr.Body.String())
}
