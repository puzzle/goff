
# 🔍 {{ .Title }}

{{ range .Files }}
## {{ .Filename }}
```diff
{{ .Diff }}
```
{{end}}
