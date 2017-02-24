package node

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/database"
)

type (
	data struct {
		database.Object `storm:"inline"`
		A               string
	}

	failDB struct {
		err         error
		SaveError   error
		OneError    error
		AllError    error
		FindError   error
		DeleteError error
	}
)

func (d *data) Validate(db database.Database) error {
	if d.A == "" {
		return errors.New("A cannot be empty")
	}

	return nil
}

func (f *failDB) Save(data interface{}) error {
	if f.SaveError != nil {
		return f.SaveError
	}

	return f.err
}

func (f *failDB) One(fieldName string, value interface{}, to interface{}) error {
	if f.OneError != nil {
		return f.OneError
	}

	return f.err
}

func (f *failDB) All(to interface{}, limit int, skip int, reverse bool) error {
	if f.AllError != nil {
		return f.AllError
	}

	return f.err
}

func (f *failDB) Find(field string, value interface{}, to interface{}, limit int, skip int, reverse bool) error {
	if f.FindError != nil {
		return f.FindError
	}

	return f.err
}

func (f *failDB) Delete(data interface{}) error {
	if f.DeleteError != nil {
		return f.DeleteError
	}

	return f.err
}

func (f *failDB) RegisterListener(listener database.Listener) {
}

var _ database.Validator = (*data)(nil)

func init() {
	database.RegisterType(data{})

	gin.SetMode(gin.ReleaseMode)
}

func TestNewRestAPI(t *testing.T) {
	db := boltdb.NewTestStore()

	r := NewRestAPI(data{}, db)

	if r == nil {
		t.Fatalf("NewRestAPI() returned nil")
	}
}

func request(db database.Database, method string, URI string, body []byte) *httptest.ResponseRecorder {
	r := NewRestAPI(data{}, db)
	router := gin.New()
	router.Use(gin.ErrorLogger())
	r.Router(router.Group("/"))

	reader := bytes.NewReader(body)
	req, _ := http.NewRequest(method, URI, reader)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	return resp
}

func TestRestApiDBFail(t *testing.T) {
	db := &failDB{err: errors.New("Fail!")}
	d := &data{A: "hejsa"}

	body, _ := json.Marshal(d)
	resp := request(db, "POST", "/", body)

	if resp.Code != 500 {
		t.Fatalf("POST / returned unexpected status code: %d", resp.Code)
	}

	resp = request(db, "GET", "/", nil)
	if resp.Code != 500 {
		t.Fatalf("GET / returned unexpected status code: %d", resp.Code)
	}

	resp = request(db, "PUT", "/id1", body)
	if resp.Code != 500 {
		t.Fatalf("PUT /id1 returned unexpected status code: %d", resp.Code)
	}

	resp = request(db, "DELETE", "/id1", nil)
	if resp.Code != 404 {
		t.Fatalf("DELETE /id1 returned unexpected status code: %d", resp.Code)
	}

	db.err = nil
	db.DeleteError = errors.New("Fail!")
	resp = request(db, "DELETE", "/id1", nil)
	if resp.Code != 500 {
		t.Fatalf("DELETE /id1 returned unexpected status code: %d", resp.Code)
	}
}

func TestRestApiCreateFailedValidation(t *testing.T) {
	db := boltdb.NewTestStore()

	d := &data{}

	body, _ := json.Marshal(d)
	resp := request(db, "POST", "/", body)

	if resp.Code != 400 {
		t.Fatalf("POST / returned unexpected status code: %d", resp.Code)
	}

	b := strings.TrimSpace(resp.Body.String())

	if b != `{"error":"A cannot be empty"}` {
		t.Fatalf("create returned unexpected body. Got '%s'", resp.Body.String())
	}
}

func TestRestApiCreateFailedJSON(t *testing.T) {
	db := boltdb.NewTestStore()

	body := []byte("this is not JSON")
	resp := request(db, "POST", "/", body)

	if resp.Code != 400 {
		t.Fatalf("POST / returned unexpected status code: %d", resp.Code)
	}

	if resp.Body.String() == "" {
		t.Fatalf("create returned empty body")
	}
}

func TestRestApiCreate(t *testing.T) {
	db := boltdb.NewTestStore()

	d := &data{A: "hejsa"}

	body, _ := json.Marshal(d)
	resp := request(db, "POST", "/", body)

	if resp.Code != 202 {
		t.Fatalf("POST / returned unexpected status code: %d", resp.Code)
	}

	if resp.Body.String() == "" {
		t.Fatalf("create returned empty body")
	}
}

func TestRestAPIList0(t *testing.T) {
	db := boltdb.NewTestStore()

	resp := request(db, "GET", "/", nil)

	if resp.Code != 200 {
		t.Fatalf("GET / returned unexpected status code: %d (Body: %s)", resp.Code, resp.Body.String())
	}

	var list []data

	err := json.Unmarshal(resp.Body.Bytes(), &list)
	if err != nil {
		t.Fatal("Get / returned invalid JSON")
	}

	if len(list) != 0 {
		t.Fatalf("GET / returned more than zero elements")
	}
}

func TestRestAPIDeleteFail(t *testing.T) {
	db := boltdb.NewTestStore()

	resp := request(db, "DELETE", "/id-that-doesnt-exist", nil)

	if resp.Code != 404 {
		t.Fatalf("GET / returned unexpected status code: %d (Body: %s)", resp.Code, resp.Body.String())
	}
}

func TestRestAPIDelete(t *testing.T) {
	db := boltdb.NewTestStore()

	d := &data{A: "hejsa"}
	db.Save(d)

	resp := request(db, "DELETE", "/"+d.ID, nil)
	if resp.Code != 202 {
		t.Fatalf("GET / returned unexpected status code: %d (Body: %s)", resp.Code, resp.Body.String())
	}
}
