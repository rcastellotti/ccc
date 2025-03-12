# CCC: curl-command-cleaner

ccc is your personal curl helper :)

<!-- It features:

- "shaking mode"
  - curl command is run once, output is saved, hashed and and compared to output of a similar curl command with (at least one) flag removed (based on some heuristics)?
- storing commands (and results) in a sqlite database
- sharing curl commands
- a web UI -->

# Usage

```bash
go run main.go <CURL command>
```

# TODO
- create a web service we control to test request functionality is not broken
- create a "magic" mode to remove common analytics cookies (intercom, GA,...)
- parse url parameters and add them with the correct option
- offer functionality to report when ccc breaks: https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/creating-an-issue#creating-an-issue-from-a-url-query
