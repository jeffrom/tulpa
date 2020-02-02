package server

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

type testCase struct {
	name  string
	files []string
	fn    func(t *testing.T)
}

func TestIntegration(t *testing.T) {
	dir, cleanup := getTempdir(t)
	defer cleanup()

	tcs := []*testCase{
		testSimple,
		testDebounce,
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			currDir, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			defer func() {
				if err := os.Chdir(currDir); err != nil {
					panic(err)
				}
			}()

			testDir := filepath.Join(dir, strings.ReplaceAll(tc.name, " ", "_"))
			t.Logf("creating dir for test: %s", testDir)
			if err := os.MkdirAll(testDir, 0755); err != nil {
				panic(err)
			}

			if err := os.Chdir(testDir); err != nil {
				panic(err)
			}
			mockCommand()
			defer resetCommand()

			for _, filePath := range tc.files {
				f, err := os.Create(filePath)
				if err != nil {
					fmt.Println(filePath, err)
				}
				defer f.Close()
			}

			tc.fn(t)
		})
	}
}

var testSimple = &testCase{
	name:  "simple",
	files: []string{"a"},
	fn: func(t *testing.T) {
		cfg, stdout, _ := newTestConfigOutErr()
		app, srv, errC := newTestCase(cfg, successHandler, "cool")

		defer app.Close()
		defer srv.Stop()
		checkNoServerError(t, errC)

		postRequest(t, srv)
		checkLinesMatch(t, stdout.String(), regexp.MustCompile("scan done in"), 1)

		touchFile(t, "a")
		postRequest(t, srv)
		checkLinesMatch(t, stdout.String(), regexp.MustCompile("scan done in"), 2)

		// fmt.Println(stdout.String())
	},
}

// func TestDebounce(t *testing.T) {
// 	testDebounce.fn(t)
// }

var testDebounce = &testCase{
	name:  "debounce",
	files: []string{"a"},
	fn: func(t *testing.T) {
		cfg, _, _ := newTestConfigOutErr()
		cfg.Debounce = 100 * time.Millisecond
		app, srv, errC := newTestCase(cfg, successHandler, "cool")

		defer app.Close()
		defer srv.Stop()
		checkNoServerError(t, errC)

		postRequest(t, srv)
		postRequest(t, srv)
		postRequest(t, srv)
		time.Sleep(150 * time.Millisecond)
		// checkLinesMatch(t, stdout.String(), regexp.MustCompile("scan done in"), 2)

		touchFile(t, "a")
		postRequest(t, srv)
		postRequest(t, srv)
		postRequest(t, srv)
		time.Sleep(150 * time.Millisecond)
		// checkLinesMatch(t, stdout.String(), regexp.MustCompile("scan done in"), 4)

		// fmt.Println(stdout)
		// fmt.Println(stdout.String())
	},
}

var successHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
})

func touchFile(t testing.TB, p string) {
	t.Helper()
	now := time.Now().Local()
	err := os.Chtimes(p, now, now)
	if err != nil {
		panic(err)
	}
}

func postRequest(t testing.TB, srv *Server) {
	t.Helper()

	uri := fmt.Sprintf("http://%s", srv.Addr())
	t.Logf("POST %s", uri)
	res, err := http.Post(uri, "text/plain", strings.NewReader("cool post"))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatal("expected 200, got", res.StatusCode)
	}
}

func checkLinesMatch(t testing.TB, s string, re *regexp.Regexp, n int) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanLines)

	matches := 0
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			matches++
		}
	}

	if n != matches {
		t.Errorf("expected %d matches for %s, but got %d", n, re.String(), matches)
	}
}

func getTempdir(t testing.TB) (string, func()) {
	dir, err := ioutil.TempDir("", "tulpa")
	if err != nil {
		panic(err)
	}

	return dir, func() {
		t.Logf("cleaning up %s", dir)
		if err := os.RemoveAll(dir); err != nil {
			fmt.Println(err)
		}
	}
}
