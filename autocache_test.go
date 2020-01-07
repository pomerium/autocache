package autocache

import (
	"net"
	"net/http/httptest"
	"testing"

	"github.com/golang/groupcache"
	"github.com/hashicorp/memberlist"
)

// todo(bdd): test coverage could be improved by using memberlist's mock
// todo(bdd): by design, groupcache's http pool panics if initialized twice

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		o          *Options
		path       string
		wantErr    bool
		wantStatus int
		wantBody   string
	}{
		{"complete config", &Options{PoolOptions: &groupcache.HTTPPoolOptions{BasePath: "/"}, SeedNodes: []string{"http://localhost:80", "http://127.0.0.1"}}, "/no_such_group/2/", false, 404, "no such group: no_such_group\n"},
		{"reject empty seed", &Options{SeedNodes: []string{}}, "", true, 0, ""},
		{"reject schemeless url", &Options{SeedNodes: []string{"localhost"}}, "", true, 0, ""},
		{"reject malformed url", &Options{SeedNodes: []string{"%^notscheme-url:80"}}, "", true, 0, ""},
		{"memberlist can't find self", &Options{SeedNodes: []string{"http://0.0.0.0"}}, "", true, 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(tt.o)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			r := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			if status := w.Code; status != tt.wantStatus {
				t.Errorf("status code: got %v want %v", status, tt.wantStatus)
			}
			if tt.wantBody != "" {
				body := w.Body.String()
				if body != tt.wantBody {
					t.Errorf("wrong body\n%s \n %s", body, tt.wantBody)
				}
			}
			ip, _, err := net.ParseCIDR("192.0.2.1/24")
			if err != nil {
				t.Fatal(err)
			}
			s.NotifyJoin(&memberlist.Node{Addr: ip})
			if len(s.peers) != 2 {
				t.Errorf("NotifyJoin failed")
			}
			s.NotifyLeave(&memberlist.Node{Addr: ip})
			if len(s.peers) != 1 {
				t.Errorf("NotifyLeave failed")
			}
			s.NotifyUpdate(&memberlist.Node{Addr: ip})
			if len(s.peers) != 1 {
				t.Errorf("NotifyUpdate failed")
			}
		})
	}
}
