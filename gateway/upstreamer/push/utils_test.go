package push

import "testing"

func Test_getTargetIdentity(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"/",
			args{
				"/",
			},
			"",
		},
		{
			"/users",
			args{
				"/users",
			},
			"users",
		},
		{
			"/users/id",
			args{
				"/users/id",
			},
			"users",
		},
		{
			"/users/id/groups",
			args{
				"/users/id/groups",
			},
			"groups",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTargetIdentity(tt.args.path); got != tt.want {
				t.Errorf("getTargetIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}
