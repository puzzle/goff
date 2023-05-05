
[[_TOC_]]

<p>
<details>
<summary><h1>🔍 {{ .Title }}</h1></summary>

# {{ .Title }}

{{ range .Files }}
## {{ .Filename }}
```diff
{{ .Diff }}
```
{{end}}

</details>
</p>