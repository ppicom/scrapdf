package scraper

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		want    string
		wantErr bool
	}{
		{
			name: "basic paragraph",
			html: `<p>Hello World</p>`,
			want: "Hello World\n\n",
		},
		{
			name: "nested elements",
			html: `<div><p>First paragraph</p><p>Second paragraph</p></div>`,
			want: "First paragraph\n\nSecond paragraph\n\n",
		},
		{
			name: "with comments",
			html: `<!-- This is a comment -->
			<p>Actual content</p>
			<!-- Another comment -->`,
			want: "Actual content\n\n",
		},
		{
			name: "with script and style",
			html: `
			<style>
				body { color: red; }
			</style>
			<script>
				console.log("test");
			</script>
			<p>Real content</p>`,
			want: "Real content\n\n",
		},
		{
			name: "list items",
			html: `<ul><li>First item</li><li>Second item</li></ul>`,
			want: "• First item\n• Second item\n\n",
		},
		{
			name: "headings",
			html: `<h1>Title</h1><p>Content</p><h2>Subtitle</h2>`,
			want: "Title\n\nContent\n\nSubtitle\n\n",
		},
		{
			name: "complex nested structure",
			html: `
			<div class="content">
				<h1>Main Title</h1>
				<!-- Navigation section -->
				<nav>
					<a href="#">Home</a>
					<a href="#">About</a>
				</nav>
				<div class="article">
					<p>First paragraph with <strong>bold</strong> text.</p>
					<ul>
						<li>Point 1</li>
						<li>Point 2</li>
					</ul>
					<p>Second paragraph.</p>
				</div>
				<style>
					.hidden { display: none; }
				</style>
				<script>
					alert("hello");
				</script>
			</div>`,
			want: "Main Title\n\nHome\nAboutFirst paragraph with\n\nboldtext.\n\n• Point 1\n• Point 2\nSecond paragraph.\n\n",
		},
		{
			name: "empty elements",
			html: `<p></p><div></div><span></span>`,
			want: "",
		},
		{
			name: "whitespace handling",
			html: `<p>  Padded   text  </p>`,
			want: "Padded   text\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stripHTMLTags(tt.html)
			if (err != nil) != tt.wantErr {
				t.Errorf("stripHTMLTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Normalize line endings for comparison
			got = strings.ReplaceAll(got, "\r\n", "\n")
			want := strings.ReplaceAll(tt.want, "\r\n", "\n")

			if got != want {
				t.Errorf("stripHTMLTags() = \n%q\nwant\n%q", got, want)
				// Print a more readable diff
				t.Errorf("\nGot (length %d):\n%s\nWant (length %d):\n%s",
					len(got), got, len(want), want)
			}
		})
	}
}

// Helper function to debug node traversal
func TestDebugNodeTraversal(t *testing.T) {
	htmlContent := `
	<div>
		<h1>Title</h1>
		<p>Paragraph</p>
		<!-- Comment -->
		<script>console.log("test");</script>
	</div>`

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	var debugNode func(*html.Node, int)
	debugNode = func(n *html.Node, depth int) {
		indent := strings.Repeat("  ", depth)
		nodeType := ""
		switch n.Type {
		case html.ElementNode:
			nodeType = "Element"
		case html.TextNode:
			nodeType = "Text"
		case html.CommentNode:
			nodeType = "Comment"
		case html.DoctypeNode:
			nodeType = "Doctype"
		}

		t.Logf("%s%s: '%s' (Data: '%s')", indent, nodeType, n.Data, strings.TrimSpace(n.Data))

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			debugNode(c, depth+1)
		}
	}

	t.Log("Node structure:")
	debugNode(doc, 0)
}
