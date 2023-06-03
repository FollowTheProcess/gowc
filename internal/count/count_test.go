package count_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/FollowTheProcess/gowc/internal/count"
)

const someText = `
Doloribus provident dolore repellat tempore iste vitae ea tempora saepe. Amet cumque perferendis perferendis. Earum earum repellendus
fuga ducimus recusandae minima molestias culpa assumenda. Ut tempora rem cum deleniti labore. Eligendi quo iste deserunt.
Harum veniam non commodi veniam excepturi.

Esse iusto velit unde non esse officiis inventore beatae. Inventore dolore totam nesciunt maiores dolore culpa iusto.
Voluptas vero in officiis fugiat illum fuga praesentium doloremque. Pariatur quisquam a quisquam illo molestiae.
Provident reprehenderit numquam veniam sapiente natus.
Exercitationem quis ut ipsa totam excepturi vel natus. Eos fuga ipsum ab aspernatur consectetur. Ad sequi vitae laboriosam.

Illo veniam occaecati error debitis inventore sed odio. Tempora incidunt praesentium placeat eius officia architecto accusamus voluptates.
`

func TestCount(t *testing.T) {
	tests := []struct {
		name string
		in   io.Reader
		want count.Result
	}{
		{
			name: "one word",
			in:   strings.NewReader("hello"),
			want: count.Result{
				Lines: 0,
				Bytes: 5,
				Words: 1,
				Chars: 5,
			},
		},
		{
			name: "two words",
			in:   strings.NewReader("hello there"),
			want: count.Result{
				Lines: 0,
				Bytes: 11,
				Words: 2,
				Chars: 11,
			},
		},
		{
			name: "empty",
			in:   strings.NewReader(""),
			want: count.Result{
				Lines: 0,
				Bytes: 0,
				Words: 0,
				Chars: 0,
			},
		},
		{
			name: "newlines",
			in:   strings.NewReader("\n\n\n"),
			want: count.Result{
				Lines: 3,
				Bytes: 3,
				Words: 0,
				Chars: 3,
			},
		},
		{
			name: "unicode",
			in:   strings.NewReader("世界 🚀 🦀\n"),
			want: count.Result{
				Lines: 1,
				Bytes: 17,
				Words: 3,
				Chars: 7,
			},
		},
		{
			name: "text",
			in:   strings.NewReader(someText),
			want: count.Result{
				Lines: 11,
				Bytes: 851,
				Words: 113,
				Chars: 851,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := count.Count(tt.in, "")
			if err != nil {
				t.Fatalf("count returned an unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nGot:\t%+v\nWanted:\t%+v\n", got, tt.want)
			}
		})
	}
}

func TestDisplay(t *testing.T) {
	count := count.Result{
		Name:  "test",
		Lines: 42,
		Bytes: 128,
		Words: 1000,
		Chars: 78926,
	}

	out := &bytes.Buffer{}
	if err := count.Display(out, false); err != nil {
		t.Fatalf("Display returned an unexpected error: %v", err)
	}

	got := out.String()
	want := "File\tBytes\tChars\tLines\tWords\ntest\t128\t78926\t42\t1000\n"

	if got != want {
		t.Errorf("\nGot:\t%#v\nWanted:\t%#v\n", got, want)
	}
}

func TestDisplayJSON(t *testing.T) {
	count := count.Result{
		Name:  "test",
		Lines: 42,
		Bytes: 128,
		Words: 1000,
		Chars: 78926,
	}

	out := &bytes.Buffer{}
	if err := count.Display(out, true); err != nil {
		t.Fatalf("Display returned an unexpected error: %v", err)
	}

	got := strings.TrimSpace(out.String())
	want := `{"name":"test","lines":42,"bytes":128,"words":1000,"chars":78926}`

	if got != want {
		t.Errorf("\nGot:\t%#v\nWanted:\t%#v\n", got, want)
	}
}

func BenchmarkCount(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}

	mobyDick := filepath.Join(cwd, "testdata", "moby_dick.txt")
	contents, err := os.ReadFile(mobyDick)
	if err != nil {
		b.Fatalf("could not read moby dick: %v", err)
	}
	r := bytes.NewReader(contents)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := count.Count(r, "bench")
		if err != nil {
			b.Fatalf("Count returned an error: %v", err)
		}
	}
}
