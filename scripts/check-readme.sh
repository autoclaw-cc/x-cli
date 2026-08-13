#!/usr/bin/env bash
# Keeps README.md / README_EN.md honest about what the repo actually contains.
#
# Two things drifted before and prompted this check:
#   - boss-cli and scholar-cli were added but never mentioned in either README
#   - both READMEs claimed "12 CLIs" long after the count reached 14
#
# Portable to macOS bash 3.2 / BSD find as well as CI.
set -uo pipefail
cd "$(dirname "$0")/.."

fail=0
note() { printf '  x %s\n' "$1"; fail=1; }

clis=$(find . -maxdepth 1 -type d -name '*-cli' | sed 's|^\./||' | sort)
count=$(printf '%s\n' "$clis" | grep -c .)
echo "found $count CLI directories"

# 1. every CLI directory is mentioned in both READMEs
for readme in README.md README_EN.md; do
  for cli in $clis; do
    grep -qF "$cli" "$readme" || note "$readme never mentions $cli"
  done
done

# 2. the stated count matches reality.
# If you reword these sentences, update the patterns here too.
check_claim() { # <file> <regex matching the number in context> <label>
  hits=$(grep -oE "$2" "$1" || true)
  if [ -z "$hits" ]; then
    note "$1: no '$3' sentence found - reworded? update scripts/check-readme.sh"
    return
  fi
  for n in $(printf '%s\n' "$hits" | grep -oE '[0-9]+'); do
    [ "$n" = "$count" ] || note "$1 says $n where the repo has $count ($3)"
  done
}

check_claim README.md    '[0-9]+ 个 CLI'      "N 个 CLI"
check_claim README.md    '这 [0-9]+ 个'        "这 N 个"
check_claim README_EN.md 'All [0-9]+ CLIs'    "All N CLIs"
check_claim README_EN.md 'All [0-9]+ in this' "All N in this repo"

if [ "$fail" = 0 ]; then echo "README is in sync"; else echo "README is out of sync with the repo"; fi
exit "$fail"
