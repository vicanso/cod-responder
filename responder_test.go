package responder

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/vicanso/cod"
)

func checkResponse(t *testing.T, resp *httptest.ResponseRecorder, code int, data string) {
	if resp.Body.String() != data ||
		resp.Code != code {
		t.Fatalf("check response fail")
	}
}

func checkJSON(t *testing.T, resp *httptest.ResponseRecorder) {
	if resp.Header().Get(cod.HeaderContentType) != cod.MIMEApplicationJSON {
		t.Fatalf("response content type should be json")
	}
}

func checkContentType(t *testing.T, resp *httptest.ResponseRecorder, contentType string) {
	if resp.Header().Get(cod.HeaderContentType) != contentType {
		t.Fatalf("response content type check fail")
	}
}

func TestResponder(t *testing.T) {
	m := NewDefaultResponder()
	req := httptest.NewRequest("GET", "https://aslant.site/", nil)

	t.Run("skip", func(t *testing.T) {
		c := cod.NewContext(nil, nil)
		done := false
		c.Next = func() error {
			done = true
			return nil
		}
		fn := NewResponder(Config{
			Skipper: func(c *cod.Context) bool {
				return true
			},
		})
		err := fn(c)
		if err != nil ||
			!done {
			t.Fatalf("skip fail")
		}
	})

	t.Run("set BodyBuffer", func(t *testing.T) {
		c := cod.NewContext(nil, nil)
		done := false
		c.Next = func() error {
			c.BodyBuffer = bytes.NewBuffer([]byte(""))
			done = true
			return nil
		}
		fn := NewResponder(Config{})
		err := fn(c)
		if err != nil ||
			!done {
			t.Fatalf("set body buffer should pass")
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		d := cod.New()
		d.Use(m)
		d.GET("/", func(c *cod.Context) error {
			return nil
		})
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		checkResponse(t, resp, 500, "category=cod-responder, message=invalid response")
	})

	t.Run("return string", func(t *testing.T) {
		d := cod.New()
		d.Use(m)
		d.GET("/", func(c *cod.Context) error {
			c.SetContentTypeByExt(".html")
			c.Body = "abc"
			return nil
		})
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		checkResponse(t, resp, 200, "abc")
		checkContentType(t, resp, "text/html; charset=utf-8")
	})

	t.Run("return bytes", func(t *testing.T) {
		d := cod.New()
		d.Use(m)
		d.GET("/", func(c *cod.Context) error {
			c.Body = []byte("abc")
			return nil
		})
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		checkResponse(t, resp, 200, "abc")
		checkContentType(t, resp, cod.MIMEBinary)
	})

	t.Run("return struct", func(t *testing.T) {
		type T struct {
			Name string `json:"name,omitempty"`
		}
		d := cod.New()
		d.Use(m)
		d.GET("/", func(c *cod.Context) error {
			c.Created(&T{
				Name: "tree.xie",
			})
			return nil
		})
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		checkResponse(t, resp, 201, `{"name":"tree.xie"}`)
		checkJSON(t, resp)
	})

	t.Run("json marshal fail", func(t *testing.T) {
		d := cod.New()
		d.Use(m)
		d.GET("/", func(c *cod.Context) error {
			c.Body = func() {}
			return nil
		})
		resp := httptest.NewRecorder()
		d.ServeHTTP(resp, req)
		checkResponse(t, resp, 500, "func() is unsupported type")
	})
}
