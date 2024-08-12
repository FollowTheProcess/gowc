package count_test

import (
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/FollowTheProcess/gowc/internal/count"
	"github.com/FollowTheProcess/test"
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

var (
	update = flag.Bool("update", false, "Update golden files")
	debug  = flag.Bool("debug", false, "Print debug output to stdout")
)

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
			got, err := count.One(tt.in, "")

			test.Ok(t, err)
			test.Equal(t, got, tt.want)
		})
	}
}

func TestCountAll(t *testing.T) {
	files := []string{
		filepath.Join("testdata", "TestCount", "moby_dick.txt"),
		filepath.Join("testdata", "TestCount", "another.txt"),
		filepath.Join("testdata", "TestCount", "onemore.txt"),
		filepath.Join("testdata", "TestCount", "dir"),
	}

	results, err := count.All(files)
	test.Ok(t, err)

	want := count.Results{
		{
			Name:  "testdata/TestCount/another.txt",
			Bytes: 608,
			Chars: 608,
			Lines: 2,
			Words: 80,
		},
		{
			Name:  "testdata/TestCount/moby_dick.txt",
			Bytes: 1232922,
			Chars: 1232922,
			Lines: 23243,
			Words: 214132,
		},
		{
			Name:  "testdata/TestCount/onemore.txt",
			Bytes: 460,
			Chars: 460,
			Lines: 2,
			Words: 63,
		},
	}

	// Sort them before comparison as order doesn't matter
	slices.SortFunc(results, cmpResult)
	slices.SortFunc(want, cmpResult)

	test.EqualFunc(t, results, want, slices.Equal)
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
	err := count.Display(out, false)
	test.Ok(t, err)

	got := out.String()
	want := filepath.Join(test.Data(t), "TestDisplay", "one.txt")

	if *debug {
		fmt.Printf("\nDEBUG (TestDisplay)\n------------\n\n%s\n", got)
	}

	if *update {
		err := os.WriteFile(want, out.Bytes(), os.ModePerm)
		test.Ok(t, err)
	}

	test.File(t, got, want)
}

func TestDisplayMultiple(t *testing.T) {
	counts := count.Results{
		{
			Name:  "one",
			Lines: 42,
			Bytes: 128,
			Words: 1000,
			Chars: 78926,
		},
		{
			Name:  "two",
			Lines: 42,
			Bytes: 128,
			Words: 1000,
			Chars: 78926,
		},
		{
			Name:  "three",
			Lines: 42,
			Bytes: 128,
			Words: 1000,
			Chars: 78926,
		},
	}

	out := &bytes.Buffer{}
	err := counts.Display(out, false)
	test.Ok(t, err)

	got := out.String()
	want := filepath.Join(test.Data(t), "TestDisplay", "all.txt")

	if *debug {
		fmt.Printf("\nDEBUG (TestDisplayMultiple)\n------------\n\n%s\n", got)
	}

	if *update {
		err := os.WriteFile(want, out.Bytes(), os.ModePerm)
		test.Ok(t, err)
	}

	test.File(t, got, want)
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
	err := count.Display(out, true)
	test.Ok(t, err)

	got := out.String()
	want := filepath.Join(test.Data(t), "TestDisplay", "one.json")

	if *debug {
		fmt.Printf("\nDEBUG (TestDisplayJSON)\n------------\n\n%s\n", got)
	}

	if *update {
		err := os.WriteFile(want, out.Bytes(), os.ModePerm)
		test.Ok(t, err)
	}

	test.File(t, got, want)
}

func TestDisplayJSONMultiple(t *testing.T) {
	counts := count.Results{
		{
			Name:  "one",
			Lines: 42,
			Bytes: 128,
			Words: 1000,
			Chars: 78926,
		},
		{
			Name:  "two",
			Lines: 42,
			Bytes: 128,
			Words: 1000,
			Chars: 78926,
		},
		{
			Name:  "three",
			Lines: 42,
			Bytes: 128,
			Words: 1000,
			Chars: 78926,
		},
	}

	out := &bytes.Buffer{}
	err := counts.Display(out, true)
	test.Ok(t, err)

	got := out.String()
	want := filepath.Join(test.Data(t), "TestDisplay", "all.json")

	if *debug {
		fmt.Printf("\nDEBUG (TestDisplayJSON)\n------------\n\n%s\n", got)
	}

	if *update {
		err := os.WriteFile(want, out.Bytes(), os.ModePerm)
		test.Ok(t, err)
	}

	test.File(t, got, want)
}

func BenchmarkCount(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}

	mobyDick := filepath.Join(cwd, "testdata", "TestCount", "moby_dick.txt")
	contents, err := os.ReadFile(mobyDick)
	if err != nil {
		b.Fatalf("could not read moby dick: %v", err)
	}
	r := bytes.NewReader(contents)

	b.ResetTimer()
	for range b.N {
		_, err := count.One(r, "bench")
		if err != nil {
			b.Fatalf("Count returned an error: %v", err)
		}
	}
}

func cmpResult(a, b count.Result) int {
	// Just compare by name
	return cmp.Compare(a.Name, b.Name)
}
