package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

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
			args: args{errors.New("oh no")},
			want: []string{"** collect-aws-metadata: oh no"},
		},
		{name: "nil",
			args: args{nil},
			want: []string{},
		},
	}

	// schedule logFatalf it to be reset after the test is finished
	origLogFatalf := logFatalf
	defer func() { logFatalf = origLogFatalf }()

	for _, tt := range tests {
		errs := []string{}

		// replace Fatalf with our error gatherer
		logFatalf = func(format string, args ...interface{}) {
			if len(args) > 0 {
				errs = append(errs, fmt.Sprintf(format, args...))
			} else {
				errs = append(errs, format)
			}
		}

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

// func Test_writeMetrics(t *testing.T) {
// 	type args struct {
// 		output   *os.File
// 		metadata *fetched_metadata
// 		prefix   string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    int
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := writeMetrics(tt.args.output, tt.args.metadata, tt.args.prefix)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("writeMetrics() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("writeMetrics() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func Test_fetchURL(t *testing.T) {
// 	type args struct {
// 		url string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []byte
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := fetchURL(tt.args.url)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("fetchURL() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("fetchURL() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func Test_fetchMetadata(t *testing.T) {
// 	type args struct {
// 		opt *collect_options
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *fetched_metadata
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := fetchMetadata(tt.args.opt)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("fetchMetadata() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("fetchMetadata() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func Test_parseArgs(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		want    *collect_options
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := parseArgs()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("parseArgs() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func Test_main(t *testing.T) {
// 	tests := []struct {
// 		name string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			main()
// 		})
// 	}
// }
//
