package entities

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	DefaultTemplate = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <title>{{ if .Title }}{{ .Title }}{{ else }}{{ .Name }}{{ end }}</title>
    <style>
      body { font-family: Arial, sans-serif; max-width: 720px; margin: 48px auto; padding: 0 16px; }
      header { display: flex; align-items: center; gap: 16px; }
      img.avatar { width: 96px; height: 96px; border-radius: 50%; object-fit: cover; }
      ul.links { list-style: none; padding: 0; }
      ul.links li { margin: 8px 0; }
    </style>
  </head>
  <body>
    <header>
      {{ if .AvatarURL }}<img class="avatar" src="{{ .AvatarURL }}" alt="{{ .Name }}"/>{{ end }}
      <div>
        <h1>{{ .Name }}</h1>
        {{ if .Headline }}<p>{{ .Headline }}</p>{{ end }}
      </div>
    </header>
    {{ if .Bio }}<p>{{ .Bio }}</p>{{ end }}
    {{ if .Links }}
      <h2>Links</h2>
      <ul class="links">
        {{ range .Links }}
          <li><a href="{{ .URL }}">{{ .Label }}</a></li>
        {{ end }}
      </ul>
    {{ end }}
  </body>
</html>
`
)

type TemplateData struct {
	Title     string
	Name      string
	Headline  string
	Bio       string
	AvatarURL string
	Links     []Link
}

type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}
