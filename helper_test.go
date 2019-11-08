package hx_test

import (
	"testing"

	"github.com/izumin5210/hx"
)

func TestPath(t *testing.T) {
	cases := []struct {
		test string
		got  string
		want string
	}{
		{
			test: "simple",
			got:  hx.Path("/api/contents", 1, "stargazers"),
			want: "/api/contents/1/stargazers",
		},
		{
			test: "stringer",
			got:  hx.Path("/api", "contents", fakeStringer("fakestringer"), "stargazers"),
			want: "/api/contents/fakestringer/stargazers",
		},
		{
			test: "abs",
			got:  hx.Path("https://api.example.com", "contents", uint32(3), "stargazers"),
			want: "https://api.example.com/contents/3/stargazers",
		},
	}

	for _, tc := range cases {
		t.Run(tc.test, func(t *testing.T) {
			if got, want := tc.got, tc.want; got != want {
				t.Errorf("hx.Path(...) returns %q, want %q", got, want)
			}
		})
	}
}
