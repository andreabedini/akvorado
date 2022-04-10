package console

import (
	"fmt"
	"io/ioutil"
	netHTTP "net/http"
	"strings"
	"testing"

	"akvorado/common/http"
	"akvorado/common/reporter"
)

func TestServeDocs(t *testing.T) {
	for _, live := range []bool{false, true} {
		name := "livefs"
		if !live {
			name = "embeddedfs"
		}
		cases := []struct {
			Path   string
			Expect string
		}{
			{"usage", `<a href=\"configuration\">configuration section</a>`},
			{"intro", `data:image/svg`},
		}
		for _, tc := range cases {
			t.Run(fmt.Sprintf("%s-%s", name, tc.Path), func(t *testing.T) {
				r := reporter.NewMock(t)
				h := http.NewMock(t, r)
				_, err := New(r, Configuration{
					ServeLiveFS: live,
				}, Dependencies{
					HTTP: h,
				})
				if err != nil {
					t.Fatalf("New() error:\n%+v", err)
				}

				resp, err := netHTTP.Get(fmt.Sprintf("http://%s/api/v0/docs/%s",
					h.Address, tc.Path))
				if err != nil {
					t.Fatalf("GET /api/v0/docs/%s:\n%+v", tc.Path, err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					t.Errorf("GET /api/v0/docs/%s: got status code %d, not 200",
						tc.Path, resp.StatusCode)
				}
				body, _ := ioutil.ReadAll(resp.Body)
				if !strings.Contains(string(body), tc.Expect) {
					t.Errorf("GET /api/v0/docs/%s: does not contain %q",
						tc.Path, tc.Expect)
				}
			})
		}
	}
}