package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// create & return a replacement for Fatalf that collects calls to Fatalf
//
// in-param *[]errs will be the strings we called Fatalf with
func genFakeFatalf(errs *[]string) func(f string, args ...interface{}) {
	return func(format string, args ...interface{}) {
		if len(args) > 0 {
			*errs = append(*errs, fmt.Sprintf(format, args...))
		} else {
			*errs = append(*errs, format)
		}
	}
}

func Test_check(t *testing.T) {
	type args struct {
		e error
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "error",
			args: args{errMissingTextfilesPath},
			want: []string{"** collect-aws-metadata: required: --textfiles-path"},
		},
		{name: "nil",
			args: args{nil},
			want: []string{},
		},
	}

	// schedule logFatalf to be reset after the test is finished
	// (logFatalf is an alias in collect.go specifically so the test can shadow it)
	origLogFatalf := logFatalf
	defer func() { logFatalf = origLogFatalf }()

	for _, tt := range tests {
		errs := []string{}
		fmt.Println(errs)

		// replace Fatalf with our error gatherer
		logFatalf = genFakeFatalf(&errs)

		t.Run(tt.name, func(t *testing.T) {
			check(tt.args.e)
			if !reflect.DeepEqual(errs, tt.want) {
				t.Errorf("check(%s) wanted %s got %s", tt.args.e, tt.want, errs)
			}
		})
	}
}

func Test_printInfo(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "hi",
			args: args{"hi"},
			want: `^collect-aws-metadata: hi$`,
		},
		{name: "empty",
			args: args{""},
			want: `^collect-aws-metadata: $`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := printInfo(tt.args.msg)
			rx := regexp.MustCompile(tt.want)
			if rx.FindStringIndex(ret) == nil {
				t.Errorf(`printInfo(%s) ! match %s`, tt.args.msg, tt.want)
			}
		})
	}
}

func Test_writeMetrics(t *testing.T) {
	type args struct {
		writer   *bytes.Buffer
		metadata *fetched_metadata
		prefix   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "none events",
			args: args{
				writer:   bytes.NewBufferString(""),
				metadata: &fetched_metadata{},
				prefix:   "hi_"},
			want:    `(?s)^hi_aws_maintenance_event_count\{instance=""\} 0$`,
			wantErr: false,
		},
		{name: "2x events",
			args: args{
				writer: bytes.NewBufferString(""),
				metadata: &fetched_metadata{instanceID: "q-qqqqqq",
					events: []maintenance_event{
						{
							EventId:   "ev-ent1",
							Code:      "system-reboot",
							NotBefore: "20 Jan 2020 09:00:43 GMT",
						}, {
							EventId:   "ev-ent2",
							Code:      "system-reboot",
							NotBefore: "20 Jan 2019 09:00:43 GMT",
						}}},
				prefix: ""},
			want:    `(?s)instance="q-qqqqqq".*\b2\b.*\bid="ev-ent1".*\b1579510843\b.*\bid="ev-ent2".*\b1547974843$`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeMetrics(tt.args.writer, tt.args.metadata, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			bb, _ := io.ReadAll(tt.args.writer)
			trimmed := strings.TrimSpace(string(bb))
			rx := regexp.MustCompile(tt.want)
			if rx.FindStringIndex(trimmed) == nil {
				t.Errorf("got %s, != %s", trimmed, tt.want)
			}
		})
	}
}

func Test_fetchURL(t *testing.T) {
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "q-qqqqqq")
	}))
	defer srv1.Close()
	srvBad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", 500)
	}))
	defer srvBad.Close()
	qs := "q-qqqqqq"
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    *string
		wantErr bool
	}{
		{name: "instance-id returns an id",
			args:    args{url: srv1.URL + "/instance-id"},
			want:    &qs,
			wantErr: false,
		},
		{name: "instance-id 404",
			args:    args{url: srvBad.URL + "/instance-id"},
			want:    nil,
			wantErr: true,
		},
		{name: "instance-id unconnectable",
			args:    args{url: "http://127.0.0.1:99999"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchURL(tt.args.url)
			strgot := strings.TrimSpace(string(got))
			if tt.want != nil && !reflect.DeepEqual(strgot, *tt.want) {
				t.Errorf("fetchURL(%s) = '%v', want '%v'", tt.args.url, strgot, *tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchURL(%s), error = '%v', wantErr '%v'", tt.args.url, err, tt.wantErr)
				return
			}
		})
	}
}

type writerFunc func(w http.ResponseWriter)

// create an httptest server with one handler that stands in for the AWS metadata service.
//
// Required parameters are two functions of type writerFunc:
// - fnID handles test calls to /1.0/meta-data/instance-id
// - fnJSON handles test calls to /latest/meta-data/events/maintenance/scheduled
//
// Returns the server with these handlers bound to write responses
func helpMakeAServer(fnID writerFunc, fnJSON writerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.String(), "/instance-id") {
				fnID(w)
			} else {
				fnJSON(w)
			}
		}))
}

func Test_fetchMetadata(t *testing.T) {
	mevs := []maintenance_event{{
		NotBefore:   "20 Jan 2019 09:00:43 GMT",
		Code:        "",
		Description: "",
		EventId:     "x-yzabc",
		NotAfter:    "",
		State:       "",
	}}
	emptyOptions := collect_options{
		textfilesPath: "",
		metricPrefix:  "",
		baseURL:       "",
	}
	tests := []struct {
		name    string
		server  *httptest.Server
		want    fetched_metadata
		wantErr bool
	}{
		{name: "good",
			server: helpMakeAServer(
				func(w http.ResponseWriter) { fmt.Fprint(w, "i-jklmn") },
				func(w http.ResponseWriter) {
					data, _ := json.Marshal(mevs)
					fmt.Fprint(w, string(data))
				},
			),
			want: fetched_metadata{
				instanceID: "i-jklmn",
				events:     mevs},
			wantErr: false},
		{name: "bad instance-id",
			server: helpMakeAServer(
				func(w http.ResponseWriter) { http.Error(w, "not my instance id", 404) },
				func(w http.ResponseWriter) { /* we don't reach this */ },
			),
			want: fetched_metadata{
				instanceID: "i-jklmn",
				events:     mevs},
			wantErr: true},
		{name: "bad JSON",
			server: helpMakeAServer(
				func(w http.ResponseWriter) { fmt.Fprint(w, "i-jklmn") },
				func(w http.ResponseWriter) { fmt.Fprint(w, "<html>Oh no you have encountered an error page</html>") },
			),
			want: fetched_metadata{
				instanceID: "i-jklmn",
				events:     mevs},
			wantErr: true},
		{name: "404 JSON",
			server: helpMakeAServer(
				func(w http.ResponseWriter) { fmt.Fprint(w, "i-jklmn") },
				func(w http.ResponseWriter) { http.Error(w, "not my JSON", 404) },
			),
			want: fetched_metadata{
				instanceID: "i-jklmn",
				events:     mevs},
			wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// assign the server url associated with this test to baseURL
			defer tt.server.Close()
			opts := emptyOptions // copy the struct
			opts.baseURL = tt.server.URL
			got, err := fetchMetadata(&opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("fetchMetadata() = %v, want %v", *got, tt.want)
			}
		})
	}
}

func Test_parseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *collect_options
		wantErr error
	}{
		{
			name:    "missing --textfiles-path should error",
			args:    []string{},
			want:    nil,
			wantErr: errMissingTextfilesPath,
		},
		{
			name: "options are set, default metricPrefix",
			args: []string{"--textfiles-path", ".", "--base-url", "http://example.com"},
			want: &collect_options{
				baseURL:       "http://example.com",
				metricPrefix:  "",
				textfilesPath: ".",
			},
			wantErr: nil,
		},
		{
			name: "options are set, use default baseURL",
			args: []string{"--textfiles-path", ".", "--metric-prefix", "asdf_"},
			want: &collect_options{
				baseURL:       "http://169.254.169.254",
				metricPrefix:  "asdf_",
				textfilesPath: ".",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if !(tt.wantErr == nil) {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseArgs() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_main(t *testing.T) {
	// fake server
	srv := helpMakeAServer(
		func(w http.ResponseWriter) { fmt.Fprintf(w, "i-jklmn") },
		func(w http.ResponseWriter) { fmt.Fprintf(w, "[]") },
	)

	// capture output and schedule to restore it at the end of the test
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() { os.Stdout = origStdout; os.Stderr = origStderr }()
	r1, w1, _ := os.Pipe()
	log.SetOutput(w1) // capture log calls too
	defer func() { log.SetOutput(origStdout) }()
	os.Stdout = w1
	os.Stderr = w1

	tests := []struct {
		name    string
		cliArgs []string
		want    string
	}{
		{name: "yes",
			cliArgs: []string{
				"collect-aws-metadata",
				"--textfiles-path=/tmp",
				"--base-url=" + srv.URL,
			},
			want: `(?s)collect-aws-metadata: Fetched http.*\b0 events.*collect-aws-metadata: Wrote /tmp/`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.cliArgs
			defer func() { os.Remove("/tmp/collect-aws-metadata.prom") }()
			main()
			w1.Close()
			rx := regexp.MustCompile(tt.want)
			got, _ := ioutil.ReadAll(r1)
			trimmed := strings.TrimSpace(string(got))
			if rx.FindStringIndex(trimmed) == nil {
				t.Errorf("Program output was `%s` wanted `%s`", got, tt.want)
			}
		})
	}
}

func TestHTTPErrorStatusCode_Error(t *testing.T) {
	type fields struct {
		url     string
		code    int
		message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name:   "printy",
		fields: fields{url: "http://example.com", code: 420, message: "420 too sick"},
		want:   "<http://example.com> 420 too sick",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &HTTPErrorStatusCode{
				url:     tt.fields.url,
				code:    tt.fields.code,
				message: tt.fields.message,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("HTTPErrorStatusCode.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
