# scholar-cli

Academic paper search CLI — parallel multi-source search, dedup, BibTeX export, and PDF download. Uses free HTTP APIs for 7 English sources and [kimi-webbridge](https://www.kimi.com/features/webbridge) for Google Scholar, CNKI, and Web of Science.

## Commands

| Command | Usage | Returns |
|---|---|---|
| `search-en` | `scholar-cli search-en --query "LLM" [--sources arxiv,semantic] [--limit 10] [--workspace DIR]` | `{papers, total, sources}` |
| `search-google` | `scholar-cli search-google --query "attention" [--limit 10] [--workspace DIR]` | `{papers, total, source}` (+ workspace stats) |
| `search-cnki` | `scholar-cli search-cnki --query "大语言模型" [--limit 20] [--workspace DIR]` | `{papers, total, source}` (+ workspace stats) |
| `search-wos` | `scholar-cli search-wos --query "transformer" [--limit 10] [--workspace DIR]` | `{papers, total, source}` (+ workspace stats) |
| `detail` | `scholar-cli detail --doi "10.1145/3442188.3445922"` | `{title, authors, abstract, ...}` |
| `download` | `scholar-cli download --doi "10.1145/..." [--output-dir .] [--scihub DOMAIN]` | `{doi, file_path, source, size_bytes}` |
| `export` | `scholar-cli export --workspace DIR [--output refs.bib]` | BibTeX to stdout or file |
| `login-status` | `scholar-cli login-status --platform wos` | `{platform, logged_in}` |

All output:

```json
{"ok": true, "data": ...}
```
or on failure (non-zero exit):

```json
{"ok": false, "error": {"code": "...", "message": "..."}}
```

## Sources

| Source | Flag name | Type | Login required |
|---|---|---|---|
| OpenAlex | `openalex` | HTTP API | No |
| Semantic Scholar | `semantic` | HTTP API | No |
| CrossRef | `crossref` | HTTP API | No |
| arXiv | `arxiv` | HTTP API | No |
| PubMed | `pubmed` | HTTP API | No |
| DBLP | `dblp` | HTTP API | No |
| bioRxiv/medRxiv | `biorxiv` | HTTP API | No |
| Google Scholar | — | WebBridge | No (CAPTCHA possible) |
| CNKI | — | WebBridge | Yes + slide CAPTCHA |
| Web of Science | — | WebBridge | Yes (institutional SSO) |

`search-en` searches all 7 HTTP API sources in parallel by default. Use `--sources` to select specific ones.

> **Note on Semantic Scholar:** S2's free anonymous tier shares a global rate-limit pool, so an occasional `count: 0` from `semantic` in the `sources` summary is normal — it doesn't mean the query failed. The other 6 sources keep working, and S2 calls auto-retry once on 429 to ride out transient saturation.

## Prerequisites

1. **Go 1.25+** for building.
2. **kimi-webbridge daemon** running on `127.0.0.1:10086` (only for Google Scholar, CNKI, WoS) — install from https://www.kimi.com/features/webbridge.
3. **CNKI**: must be logged in at https://www.cnki.net and have passed the slide CAPTCHA at https://kns.cnki.net.
4. **WoS**: must be logged in via institutional VPN/SSO.

## Build

```bash
go build -o scholar-cli .
./scholar-cli --help
```

## Quick test

```bash
# English multi-source search (no login needed)
./scholar-cli search-en --query "large language model" --limit 5

# Specify sources
./scholar-cli search-en --query "CRISPR" --sources pubmed,biorxiv --limit 10

# Enrich by DOI
./scholar-cli detail --doi "10.1145/3586183.3606763"

# Download PDF
./scholar-cli download --doi "10.48550/arxiv.2310.08560" --output-dir ./papers

# Workspace + BibTeX export
./scholar-cli search-en --query "transformer" --workspace ./my-survey --limit 10
./scholar-cli search-google --query "transformer" --workspace ./my-survey --limit 10
./scholar-cli export --workspace ./my-survey --output refs.bib
```

## Download channels

PDF download tries sources in order: arXiv direct → open access pdf_url → Unpaywall API.

Sci-Hub is **disabled by default** — Sci-Hub access is under court injunctions in several jurisdictions. To opt in, pass `--scihub <domain>` (e.g. `--scihub sci-hub.se`); use only where legal in your jurisdiction. Without the flag, the cascade stops at Unpaywall.

## Contact email for polite-pool APIs

CrossRef, OpenAlex, NCBI E-utilities, and Unpaywall ask requesting clients to identify themselves with a contact email. By default scholar-cli sends `scholar-cli@example.com`. To use your own (recommended if you query heavily), set:

```bash
export SCHOLAR_CLI_EMAIL="you@example.org"
```

## Layout

```
scholar-cli/
├── main.go
├── browser/client.go        # kimi-webbridge HTTP client
├── output/output.go         # JSON contract: {ok, data} / {ok, error}
├── paper/
│   ├── paper.go             # Paper struct, dedup, merge
│   └── bibtex.go            # BibTeX generation
├── search/
│   ├── search.go            # Parallel multi-source orchestrator
│   ├── enrich.go            # DOI enrichment (CrossRef + S2)
│   ├── http.go              # Shared HTTP helpers
│   ├── openalex.go          # OpenAlex API
│   ├── semantic.go          # Semantic Scholar API
│   ├── crossref.go          # CrossRef API
│   ├── arxiv.go             # arXiv API (XML)
│   ├── pubmed.go            # PubMed E-utilities (XML)
│   ├── dblp.go              # DBLP API
│   ├── biorxiv.go           # bioRxiv via CrossRef prefix filter
│   ├── google.go            # Google Scholar via WebBridge
│   ├── cnki.go              # CNKI via WebBridge
│   └── wos.go               # Web of Science via WebBridge
├── store/store.go           # Workspace: papers.json persistence
├── download/download.go     # PDF download (multi-channel)
├── cmd/root.go              # Cobra commands + flags
└── README.md
```

## License

MIT (see `LICENSE` — to be added).
