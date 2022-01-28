package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestShouldGetPosts(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")

	// create app with mocked db, request and response to test
	app := &api{db}
	req, err := http.NewRequest("GET", "http://localhost/posts", nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected while creating request", err)
	}
	w := httptest.NewRecorder()

	// перед выполнением http функции настраиваем ожидаемое поведение нашего драйвера по работе с БД
	rows := sqlmock.NewRows([]string{"id", "title", "views"}).
		AddRow("1", "name1", 5).
		AddRow("2", "name2", 15)

	mock.ExpectQuery("^SELECT (.+) FROM videos ORDER BY views LIMIT").WillReturnRows(rows)

	// выполняем тестируемую функцию
	app.videos(w, req)

	if w.Code != 200 {
		t.Fatalf("expected status code to be 200, but got: %d", w.Code)
	}

	type resp struct {
		Videos []Video
	}

	expected := resp{
		Videos: []Video{
			{Id: "1", Title: "name1", Views: 5},
			{Id: "2", Title: "name2", Views: 15},
		}}

	var actual resp

	err = json.Unmarshal(w.Body.Bytes(), &actual)
	if err != nil {
		t.Fatalf("error while unmarshal response body: %s", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("the expected json: %+v is different from actual %+v", expected, actual)
	}

	// проверяем что произошло все что ожидали
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
