package gotion_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/sidc9/gotion"
	"github.com/sidc9/gotion/filter"
)

var saveResponse bool

func TestMain(m *testing.M) {
	flag.BoolVar(&saveResponse, "save", false, "call the real api and save the response")
	flag.Parse()

	if saveResponse {
		fmt.Println("-> Calling real api and saving response(s)")
	}

	code := m.Run()
	os.Exit(code)
}

func TestListDatabases(t *testing.T) {
	is := is.New(t)

	respOut := filepath.Join("testdata", "list_db.txt")
	var req *http.Request
	c := setup(t, respOut, req)
	resp, err := c.ListDatabases()

	is.NoErr(err)
	is.Equal(resp.Object, "list")
	is.Equal(resp.HasMore, false)
	is.Equal(resp.NextCursor, "")

	is.Equal(len(resp.Results), 1)

	item := resp.Results[0]
	is.Equal(item.Object, "database")
}

func TestGetDatabase(t *testing.T) {
	is := is.New(t)

	t.Run("returns error if id is empty", func(t *testing.T) {
		c := &gotion.Client{}
		db, err := c.GetDatabase("")
		is.Equal(err, fmt.Errorf("id is required"))
		is.Equal(db, nil)
	})

	respOut := filepath.Join("testdata", "get_db.txt")
	var req *http.Request
	c := setup(t, respOut, req)

	db, err := c.GetDatabase("934c6132-4ea7-485e-9b0d-cf1a083e0f3f")
	is.NoErr(err)
	is.Equal(db.Object, "database")
	is.Equal(db.ID, "934c6132-4ea7-485e-9b0d-cf1a083e0f3f")
	is.Equal(db.Title[0].PlainText, "API Testing")
	is.Equal(db.Title[0].Text.Content, "API Testing")
	is.Equal(db.Properties["age"].Number.Format, "number")
	is.Equal(db.Properties["gender"].Select.Options[0].Name, "male")
	is.Equal(db.Properties["gender"].Select.Options[1].Name, "female")
	is.Equal(db.Properties["hobbies"].MultiSelect.Options[0].Name, "reading")
	is.Equal(db.Properties["hobbies"].MultiSelect.Options[1].Name, "cycling")
	is.Equal(db.Properties["hobbies"].MultiSelect.Options[2].Name, "swimming")
	is.Equal(db.Properties["description"].RichText, struct{}{})
	is.Equal(db.Properties["created at"].RichText, struct{}{})
}

func TestQueryDatabase(t *testing.T) {
	t.Run("returns error if id is empty", func(t *testing.T) {
		is := is.New(t)
		c := &gotion.Client{}
		db, err := c.QueryDatabase("", nil)
		is.Equal(err, fmt.Errorf("id is required"))
		is.Equal(db, nil)
	})

	t.Run("simple query", func(t *testing.T) {
		is := is.NewRelaxed(t)

		respOut := filepath.Join("testdata", "query_db.txt")
		var req *http.Request
		c := setup(t, respOut, req)

		pages, err := c.QueryDatabase("934c6132-4ea7-485e-9b0d-cf1a083e0f3f", nil)
		is.NoErr(err)
		is.Equal(len(pages.Results), 2)

		if req != nil {
			is.Equal(req.Method, http.MethodPost)
			is.Equal(req.URL.String(), "/databases/934c6132-4ea7-485e-9b0d-cf1a083e0f3f/query")
		}
	})

	t.Run("with number filter", func(t *testing.T) {
		is := is.NewRelaxed(t)

		respOut := filepath.Join("testdata", "query_db_with_filter.txt")
		var req *http.Request
		c := setup(t, respOut, req)

		filt := filter.NewNumberFilter("age").GreaterThan(24)
		q := gotion.NewDBQuery().WithFilter(filt)

		pages, err := c.QueryDatabase("934c6132-4ea7-485e-9b0d-cf1a083e0f3f", q)
		is.NoErr(err)
		is.Equal(len(pages.Results), 1)
		is.Equal(pages.Results[0].Properties["age"].Number, 25)
		is.Equal(pages.Results[0].ID, "63d396a6-9687-4ea6-80c8-eea4c6212658")

		if req != nil {
			is.Equal(req.Method, http.MethodPost)
			is.Equal(req.URL.String(), "/databases/934c6132-4ea7-485e-9b0d-cf1a083e0f3f/query")
		}
	})

	t.Run("with compound filter", func(t *testing.T) {
		is := is.NewRelaxed(t)

		respOut := filepath.Join("testdata", "query_db_with_compound_filter.txt")
		var req *http.Request
		c := setup(t, respOut, req)

		f1 := filter.NewNumberFilter("age").Equals(23)
		f2 := filter.NewTextFilter("description").Contains("mary")
		ff, err := filter.NewOrFilter(f1, f2)
		is.NoErr(err)

		q := gotion.NewDBQuery().WithFilter(ff)

		pages, err := c.QueryDatabase("934c6132-4ea7-485e-9b0d-cf1a083e0f3f", q)
		is.NoErr(err)
		is.Equal(len(pages.Results), 2)
		is.Equal(pages.Results[0].ID, "63d396a6-9687-4ea6-80c8-eea4c6212658")
		is.Equal(pages.Results[1].ID, "a0e3feca-85c9-440f-91cc-8c367d6aa9f4")

		if req != nil {
			is.Equal(req.Method, http.MethodPost)
			is.Equal(req.URL.String(), "/databases/934c6132-4ea7-485e-9b0d-cf1a083e0f3f/query")
		}
	})

	t.Run("with sort", func(t *testing.T) {
		is := is.NewRelaxed(t)

		respOut := filepath.Join("testdata", "query_db_with_sort.txt")
		var req *http.Request
		c := setup(t, respOut, req)

		sort := gotion.NewPropertySort("age", gotion.SortAscending)
		q := gotion.NewDBQuery().WithSorts([]*gotion.Sort{sort})

		pages, err := c.QueryDatabase("934c6132-4ea7-485e-9b0d-cf1a083e0f3f", q)
		is.NoErr(err)
		is.Equal(len(pages.Results), 2)

		item1 := pages.Results[0]
		item2 := pages.Results[1]

		is.Equal(item1.Properties["name"].Title[0].PlainText, "john")
		is.Equal(item2.Properties["name"].Title[0].PlainText, "mary")
		is.Equal(item1.Properties["age"].Number, 23)
		is.Equal(item2.Properties["age"].Number, 25)
		is.Equal(item1.Properties["description"].RichText[0].PlainText, "his name is john")
		is.Equal(item1.Properties["description"].RichText[0].Text.Content, "his name is john")
		is.Equal(item2.Properties["description"].RichText[0].PlainText, "her name is mary")
		is.Equal(item2.Properties["description"].RichText[0].Text.Content, "her name is mary")
		is.Equal(item1.Properties["gender"].Select.Name, "male")
		is.Equal(item2.Properties["gender"].Select.Name, "female")
		is.Equal(item1.Properties["hobbies"].MultiSelect[0].Name, "reading")
		is.Equal(item1.Properties["hobbies"].MultiSelect[1].Name, "cycling")
		is.Equal(item2.Properties["hobbies"].MultiSelect[0].Name, "cycling")
		is.Equal(item2.Properties["hobbies"].MultiSelect[1].Name, "swimming")
		is.Equal(item1.Properties["is admin"].Checkbox, false)
		is.Equal(item2.Properties["is admin"].Checkbox, true)
		is.Equal(item1.Properties["created at"].CreatedTime, "2021-05-14T09:09:49.526Z")
		is.Equal(item2.Properties["created at"].CreatedTime, "2021-05-15T13:28:00.000Z")

		if req != nil {
			is.Equal(req.Method, http.MethodPost)
			is.Equal(req.URL.String(), "/databases/934c6132-4ea7-485e-9b0d-cf1a083e0f3f/query")
		}
	})
}

func setup(t *testing.T, saveRespFilename string, req *http.Request) *gotion.Client {
	is := is.New(t)

	var c *gotion.Client

	apiKey, err := loadAPIKey()
	is.NoErr(err)

	if saveResponse {
		t.Logf("    * saving to: %s\n", saveRespFilename)
		c = gotion.NewClient(apiKey, "")
		c.SaveResponse(saveRespFilename)
	} else {
		h := func(w http.ResponseWriter, r *http.Request) {
			// instead of asserting the request fields here,
			// save the request and assert in the main test
			req = r

			b, err := ioutil.ReadFile(saveRespFilename)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(b)
		}

		srv := httptest.NewServer(http.HandlerFunc(h))
		t.Cleanup(srv.Close)
		c = gotion.NewClient(apiKey, srv.URL)
	}

	return c
}

func loadAPIKey() (string, error) {
	b, err := ioutil.ReadFile(".env")
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(b), "\n"), nil
}
