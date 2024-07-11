package serverify

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

type testReqLog struct {
	Method  string
	Path    string
	Headers map[string]string
	Query   map[string]string
	Body    string
}

type testHandler struct {
	statusCode int
	body       string
	requests   []testReqLog
	ts         *httptest.Server
}

func newServerify(statusCode int, body string) (*Serverify, *testHandler) {
	h := &testHandler{
		statusCode: statusCode,
		body:       body,
		requests:   []testReqLog{},
	}
	ts := httptest.NewServer(h)
	h.ts = ts

	return New(ts.URL), h
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	toSingleStringMap := func(m map[string][]string) map[string]string {
		s := make(map[string]string)
		for k, v := range m {
			s[k] = v[0]
		}
		return s
	}

	body, _ := io.ReadAll(r.Body)

	h.requests = append(h.requests, testReqLog{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: toSingleStringMap(r.Header),
		Query:   toSingleStringMap(r.URL.Query()),
		Body:    string(body),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(h.statusCode)
	fmt.Fprint(w, h.body)
}

func (h *testHandler) Close() {
	h.ts.Close()
}

func id(index int, _ any) string {
	return strconv.Itoa(index)
}

var _ = Describe("Serverify", func() {
	Describe("CreateSession()", func() {
		It("returns a new session when succeeds", func() {
			s, h := newServerify(202, `{"session":"test_session"}`)
			defer h.Close()

			session, err := s.CreateSession("test_session")

			Expect(err).NotTo(HaveOccurred())
			Expect(session.Name()).To(Equal("test_session"))
			Expect(h.requests).To(MatchAllElementsWithIndex(id, Elements{
				"0": MatchAllFields(Fields{
					"Method":  Equal("POST"),
					"Path":    Equal("/session"),
					"Headers": HaveKeyWithValue("Content-Type", "application/json"),
					"Query":   BeEmpty(),
					"Body":    MatchJSON(`{"session":"test_session"}`),
				}),
			}))
		})

		It("returns error when request fails", func() {
			s, h := newServerify(409, `{"serverify_error":{"message":"session \"test_session\" is already exists"}}`)
			defer h.Close()

			_, err := s.CreateSession("test_session")

			Expect(err).To(Equal(Error{
				StatusCode: 409,
				ServerifyError: struct{ Message string }{
					Message: `session "test_session" is already exists`,
				},
			}))
			Expect(h.requests).To(MatchAllElementsWithIndex(id, Elements{
				"0": MatchAllFields(Fields{
					"Method":  Equal("POST"),
					"Path":    Equal("/session"),
					"Headers": HaveKeyWithValue("Content-Type", "application/json"),
					"Query":   BeEmpty(),
					"Body":    MatchJSON(`{"session":"test_session"}`),
				}),
			}))
		})
	})
})

var _ = Describe("Session", func() {
	Describe("Logs()", func() {
		It("returns logs", func() {
			s, h := newServerify(200, `{"histories": [{"method": "POST", "headers": {"Content-Type": "application/json"}, "path": "/test", "query": {"qk": "qv"}, "body": "[1, 2]", "requestedAt": "2024-07-01T15:00:00Z"}]}`)
			defer h.Close()
			session := &Session{name: "test_session", serverify: s}

			logs, err := session.Logs()

			Expect(err).NotTo(HaveOccurred())
			requestedAt, _ := time.Parse(time.RFC3339, "2024-07-01T15:00:00Z")
			Expect(logs).To(Equal(&Logs{
				Histories: []Log{
					{
						Method:      "POST",
						Path:        "/test",
						Headers:     map[string]string{"Content-Type": "application/json"},
						Query:       map[string]string{"qk": "qv"},
						Body:        "[1, 2]",
						RequestedAt: requestedAt,
					},
				},
			}))
		})
	})

	Describe("Delete()", func() {
		It("returns nil when succeeds", func() {
			s, h := newServerify(200, `{"session": "test_session"}`)
			defer h.Close()
			session := &Session{name: "test_session", serverify: s}

			err := session.Delete()

			Expect(err).NotTo(HaveOccurred())
			Expect(h.requests).To(MatchAllElementsWithIndex(id, Elements{
				"0": MatchFields(IgnoreExtras, Fields{
					"Method": Equal("DELETE"),
					"Path":   Equal("/session/test_session"),
					"Query":  BeEmpty(),
				}),
			}))
		})

		It("returns error when fails", func() {
			s, h := newServerify(404, `{"serverify_error":{"message":"session \"test_session\" is not found"}}`)
			defer h.Close()
			session := &Session{name: "test_session", serverify: s}

			err := session.Delete()

			Expect(err).To(Equal(Error{
				StatusCode: 404,
				ServerifyError: struct{ Message string }{
					Message: `session "test_session" is not found`,
				},
			}))
			Expect(h.requests).To(MatchAllElementsWithIndex(id, Elements{
				"0": MatchFields(IgnoreExtras, Fields{
					"Method": Equal("DELETE"),
					"Path":   Equal("/session/test_session"),
					"Query":  BeEmpty(),
				}),
			}))
		})
	})

	Describe("BaseURL()", func() {
		It("returns base URL for the session", func() {
			s, h := newServerify(404, `{"serverify_error":{"message":"session \"test_session\" is not found"}}`)
			defer h.Close()
			session := &Session{name: "test_session", serverify: s}

			Expect(session.BaseURL()).To(Equal(h.ts.URL + "/mock/test_session"))
		})
	})
})
