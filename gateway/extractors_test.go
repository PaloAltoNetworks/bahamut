package gateway

import (
	"net/http"
	"testing"
)

func Test_defaultSourceExtractor_ExtractSource(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		f       defaultSourceExtractor
		args    args
		want    string
		wantErr bool
	}{
		{
			"no header",
			defaultSourceExtractor{},
			args{
				&http.Request{Header: http.Header{}},
			},
			"default",
			false,
		},
		{
			"header",
			defaultSourceExtractor{},
			args{
				&http.Request{Header: http.Header{"Authorization": {"Bearer X"}}},
			},
			"763578776710062675",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := defaultSourceExtractor{}
			got, err := f.ExtractSource(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultSourceExtractor.ExtractSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("defaultSourceExtractor.ExtractSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultTCPSourceExtractor_ExtractSource(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		f       defaultTCPSourceExtractor
		args    args
		want    string
		wantErr bool
	}{
		{
			"source",
			defaultTCPSourceExtractor{},
			args{
				&http.Request{RemoteAddr: "1.1.1.1"},
			},
			"1.1.1.1",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := defaultTCPSourceExtractor{}
			got, err := f.ExtractSource(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultTCPSourceExtractor.ExtractSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("defaultTCPSourceExtractor.ExtractSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
