# A literature review in one sentence

English | [中文](./review-rag-literature.md)

## The scenario

You need to get oriented in a research area you don't know. The manual version: search Google Scholar, search arXiv again, strip the duplicates by hand, open each result to read the abstract, save the PDFs you can get, and then retype it all as BibTeX. That's a morning gone, and you start from zero on the next topic.

This recipe hands the whole chain to your agent — search, dedup, enrich, download, export — from a single description.

**No account required for any of it.** arXiv, Semantic Scholar, CrossRef and OpenAlex are open APIs: no signup, no API key.

## Prerequisites

- Download and install scholar-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)
- Install the skill: `npx skills add better-world-ai/x-cli --skill paper-research`
- Optional: `export SCHOLAR_CLI_EMAIL=you@example.org` — CrossRef, OpenAlex and Unpaywall run "polite pools" and give steadier quota to requests carrying a contact address. **Unpaywall rejects the default placeholder outright**, so without this you lose one PDF download route.

If you stick to the open sources, you don't even need kimi-webbridge. It's only required to add Google Scholar (see below).

## Send this to Claude

```
Do a literature sweep on RAG evaluation for me.

Search arXiv, Semantic Scholar and CrossRef for "retrieval augmented
generation evaluation", 8 per source, into a workspace at ./rag-review,
merging duplicates.

Show me the list first — title, year, source, whether it has a DOI.
Once I pick a few, enrich those, download whatever PDFs are openly
available into ./rag-review/pdfs, and export the whole workspace as BibTeX.
```

## What you get back

**Step 1 — cross-source search with automatic dedup.** Eight per source, 16 unique papers land in the workspace:

```
papers_added: 16
sources: [{'name': 'arxiv', 'count': 8},
          {'name': 'semantic', 'count': 0, 'error': '...rate limited (HTTP 429) after 4 attempts'},
          {'name': 'crossref', 'count': 8}]
```

Note the middle line: **when a source is throttled it says so**, and the other two carry on. You get 16 papers plus an honest note, not an empty result dressed up as success.

**Step 2 — enrich a single paper.**

```
title    : MODE-RAG: Manifold Outlier Diagnosis and Energy-based Retrie...
year     : 2026 | venue: Proceedings of the 2nd Workshop on Multi...
authors  : Zehang Wei, JiaXin Dai, Jiamin Yan
sources  : ['crossref', 'semantic']
```

arXiv papers resolve too — it goes to arXiv's own API for those:

```
title  : Retrieval-Augmented Generation for Knowledge-Intensive NLP T...
sources: ['arxiv'] | year: 2020
```

**Step 3 — download open-access PDFs.**

```
arxiv | 885323 bytes | Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks.pdf
```

**Step 4 — export BibTeX.** All 16 at once, ready to drop into LaTeX:

```bibtex
@misc{qi2025arrag,
  title = {AR-RAG: Autoregressive Retrieval Augmentation for Image Generation},
  author = {Jingyuan Qi and Zhiyang Xu and Qifan Wang and Lifu Huang},
  year = {2025},
  url = {https://arxiv.org/pdf/2506.06962v3},
  eprint = {2506.06962},
  archiveprefix = {arXiv},
}
```

The workspace accumulates. Point repeated searches on different keywords at the same directory and duplicates merge on their own, then export once at the end.

## What this uses

[scholar-cli](../scholar-cli/) + the [paper-research](../skills/paper-research/) skill

## Adding Google Scholar (optional)

Google Scholar's edge is **citation counts**, which the open APIs don't provide:

```
· Ragas: Automated evaluation of retrieval augmented g...
    S Es, J James, LE Anke | Proceedings of the 18th … | 2024 | cites: 2041
```

Two extra prerequisites:

1. Install [kimi-webbridge](https://www.kimi.com/features/webbridge) — it drives your local Chrome
2. Google Scholar **may show a CAPTCHA**. Solve it in the browser yourself, then ask the agent to retry

Add "search Google Scholar too, I want citation counts" to the prompt.

## The two sources that need institutional access

`scholar-cli` also covers CNKI and Web of Science, but both require a login: CNKI needs you signed in with its slider CAPTCHA cleared, WoS needs institutional VPN or SSO. If you have either, just say "search CNKI as well" in the prompt. **Without them, nothing above changes.**

## Make it yours

Swap in your own topic. A few variations you can just say out loud:

- **Recent only**: "only papers from 2024 onward"
- **Full text only**: "skip anything without a downloadable PDF"
- **Sweep in rounds**: rerun with new keywords against the same workspace — dedup is automatic
- **Chinese literature**: with CNKI access, add "use CNKI for the Chinese-language ones"
