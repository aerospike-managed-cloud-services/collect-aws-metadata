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
				t.Errorf(`PrintInfo(%s) ! match %s`, tt.args.msg, tt.want)
			}
		})
	}
}

// func TestWriteMetrics(t *testing.T) {
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
// 			got, err := WriteMetrics(tt.args.output, tt.args.metadata, tt.args.prefix)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("WriteMetrics() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("WriteMetrics() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func TestFetchURL(t *testing.T) {
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
// 			got, err := FetchURL(tt.args.url)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("FetchURL() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("FetchURL() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func TestFetchMetadata(t *testing.T) {
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
// 			got, err := FetchMetadata(tt.args.opt)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("FetchMetadata() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("FetchMetadata() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func TestParseArgs(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		want    *collect_options
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := ParseArgs()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ParseArgs() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ParseArgs() = %v, want %v", got, tt.want)
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
