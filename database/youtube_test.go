package database

import "testing"

func TestParseYouTubeVideoId(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"dQw4w9wgXcQ", "dQw4w9wgXcQ"},
		{"https://www.youtube.com/watch?v=dQw4w9wgXcQ", "dQw4w9wgXcQ"},
		{"https://youtu.be/dQw4w9wgXcQ", "dQw4w9wgXcQ"},
		{"youtube.com/watch?v=dQw4w9wgXcQ&t=12s", "dQw4w9wgXcQ"},
		{"https://www.youtube.com/embed/dQw4w9wgXcQ", "dQw4w9wgXcQ"},
	}
	for _, c := range cases {
		got, err := ParseYouTubeVideoId(c.in)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q: got %q want %q", c.in, got, c.want)
		}
	}
	if _, err := ParseYouTubeVideoId("nope"); err == nil {
		t.Errorf("expected error for invalid input")
	}
}
