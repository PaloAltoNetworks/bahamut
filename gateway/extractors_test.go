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
			NewDefaultSourceExtractor("").(defaultSourceExtractor),
			args{
				&http.Request{Header: http.Header{}},
			},
			"default",
			false,
		},
		{
			"header",
			NewDefaultSourceExtractor("").(defaultSourceExtractor),
			args{
				&http.Request{Header: http.Header{"Authorization": {"Bearer X"}}},
			},
			"763578776710062675",
			false,
		},
		{
			"cookie",
			NewDefaultSourceExtractor("X-Token").(defaultSourceExtractor),
			args{
				func() *http.Request {
					r := &http.Request{Header: http.Header{}}
					r.AddCookie(&http.Cookie{Name: "X-Token", Value: "X"})
					return r
				}(),
			},
			"15784302077936868069",
			false,
		},
		{
			"cookie with no configured authCookieName",
			NewDefaultSourceExtractor("").(defaultSourceExtractor),
			args{
				func() *http.Request {
					r := &http.Request{Header: http.Header{}}
					r.AddCookie(&http.Cookie{Name: "X-Token", Value: "X"})
					return r
				}(),
			},
			"default",
			false,
		},
		{
			"empty cookie",
			NewDefaultSourceExtractor("X-Token").(defaultSourceExtractor),
			args{
				func() *http.Request {
					r := &http.Request{Header: http.Header{}}
					r.AddCookie(&http.Cookie{})
					return r
				}(),
			},
			"default",
			false,
		},
		{
			"cookie and header (cookie should have prio)",
			NewDefaultSourceExtractor("X-Token").(defaultSourceExtractor),
			args{
				func() *http.Request {
					r := &http.Request{Header: http.Header{"Authorization": {"Bearer X"}}}
					r.AddCookie(&http.Cookie{Name: "X-Token", Value: "X"})
					return r
				}(),
			},
			"15784302077936868069",
			false,
		},
		{
			"empty cookie and header (header should fall back)",
			NewDefaultSourceExtractor("X-Token").(defaultSourceExtractor),
			args{
				func() *http.Request {
					r := &http.Request{Header: http.Header{"Authorization": {"Bearer X"}}}
					r.AddCookie(&http.Cookie{})
					return r
				}(),
			},
			"763578776710062675",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.f
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
