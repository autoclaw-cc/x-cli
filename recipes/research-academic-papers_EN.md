# Research academic papers in one prompt

English | [中文](./research-academic-papers.md)

## When to use

You are preparing a proposal, literature review, or quick survey of a research area and do not want to repeat the same search, metadata cleanup, PDF download, and citation-export work across several databases. Let the AI turn those steps into one inspectable research workflow.

## Setup

- Download and install scholar-cli from [Releases](https://github.com/better-world-ai/x-cli/releases)
- Install the skill: `npx skills add better-world-ai/x-cli --skill paper-research`
- [kimi-webbridge](https://www.kimi.com/features/webbridge) is only needed for Google Scholar, CNKI, or Web of Science; CNKI and WoS also require a browser login

## Send to Claude

```
Research “evaluation methods for retrieval-augmented generation (RAG)” and keep the workspace at ./research/rag-evaluation.

Start with scholar-cli search-en across the arxiv and semantic sources, save 15 results to the workspace, and select the 5 most relevant papers that have a DOI. Run detail --doi for each selected paper to enrich its authors, abstract, year, and citation data. Run download for papers with a legal open-access channel and save the PDFs under ./research/rag-evaluation/papers; do not enable --scihub. Finally, run export on the workspace and write the BibTeX file to ./research/rag-evaluation/refs.bib.

Return a table of the core papers, explain why each is worth reading, report whether its PDF was downloaded, and list failed sources or missing DOIs. Do not present unverified details as facts.
```

## What you'll get

- A deduplicated literature workspace that can be extended with later searches
- Complete metadata and reading rationale for 5 core papers
- Legally open-access PDFs, with explicit reasons for anything not downloaded
- A `refs.bib` file ready for Zotero, LaTeX, and other citation tools
- A visible list of source failures, rate limits, and missing DOIs instead of silently dropped results

## Uses

[scholar-cli](../scholar-cli/) + [paper-research](../skills/paper-research/) skill

## Adapt

Replace RAG with your topic and tune `--sources` for the field. For Chinese literature, add `search-cnki` after signing in and completing any slider verification in the browser. For Web of Science, run `login-status --platform wos` first to confirm institutional access, then search with `search-wos`. Change the paper count, workspace, or export path as needed, but do not add `--scihub` to the default workflow.
