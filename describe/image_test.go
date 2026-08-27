package describe

import (
	"testing"

	corepb "github.com/projecteru2/core/rpc/gen"
)

func TestImages(t *testing.T) {
	msgs := []*corepb.ListImageMessage{
		{
			Nodename: "node1",
			Images:   []*corepb.ImageItem{{Id: "sha256:aaa", Tags: []string{"app:latest", "app:v1"}}},
		},
	}

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{name: "table", format: "", want: `┌───────┬────────────┬────────────┐
│ NODE  │ IMAGE      │ TAGS       │
├───────┼────────────┼────────────┤
│ node1 │ sha256:aaa │ app:latest │
├───────┼────────────┼────────────┤
│       │            │ app:v1     │
└───────┴────────────┴────────────┘
`},
		{name: "json", format: "json", want: `[
  {
    "images": [
      {
        "id": "sha256:aaa",
        "tags": [
          "app:latest",
          "app:v1"
        ]
      }
    ],
    "nodename": "node1"
  }
]
`},
		{name: "yaml", format: "yaml", want: `- images:
  - id: sha256:aaa
    tags:
    - app:latest
    - app:v1
  nodename: node1

`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Format = tt.format
			t.Cleanup(func() { Format = "" })

			if got := captureStdout(t, func() { Images(msgs...) }); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestImagesWithoutAnyImage(t *testing.T) {
	tests := []struct {
		name string
		msgs []*corepb.ListImageMessage
	}{
		{name: "no node reported"},
		{name: "node reported no image", msgs: []*corepb.ListImageMessage{{Nodename: "node1"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := captureStdout(t, func() { Images(tt.msgs...) }); got != "no images\n" {
				t.Errorf("got %q, want %q", got, "no images\n")
			}
		})
	}
}
