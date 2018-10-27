package bahamut

import "testing"

func Test_sanitizeURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test /toto",
			args{
				"/toto",
			},
			"/toto",
		},
		{
			"test /v/1/toto",
			args{
				"/v/1/toto",
			},
			"/toto",
		},
		{
			"test /toto/xxxxxxx",
			args{
				"/toto/xxxxxxx",
			},
			"/toto/:id",
		},
		{
			"test /v/1/toto/xxxxxxx",
			args{
				"/v/1/toto/xxxxxxx",
			},
			"/toto/:id",
		},
		{
			"test /toto/xxxxxxx/titi",
			args{
				"/toto/xxxxxxx/titi",
			},
			"/toto/:id/titi",
		},
		{
			"test /v/1/toto/xxxxxxx/titi",
			args{
				"/v/1/toto/xxxxxxx/titi",
			},
			"/toto/:id/titi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeURL(tt.args.url); got != tt.want {
				t.Errorf("sanitizeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
