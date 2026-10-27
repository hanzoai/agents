{
  "summary": "Fix 103 broken relative links in the README",
  "evidence": "Checked at 6ab6811",
  "details": "## Summary\n103 relative links in `README.md` point to `agents/`, `workflows/` and `tools/` folders that do not exist at the repo root, so every agent link on GitHub is a 404. The files are under `.hanzo/agents/`.",
  "file": "README.md",
  "status": "beta",
  "license": "MIT"
}
