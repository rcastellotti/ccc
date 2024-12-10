# CCC: curl-command-cleaner

ccc is your personal curl helper :)

It features:

* "shaking mode"
    + curl command is run once, output is saved, hashed and and compared to output of a similar curl command with (at least one) flag removed (based on some heuristics)?
* storing commands (and results) in a sqlite database
* sharing curl commands
* a web UI

# Usage

```bash
    go run main.go <CURL command>
```
# TODO
Parse analtytics cookies and send some gibberish

+ https://www.intercom.com/help/en/articles/2361922-intercom-messenger-cookies
    + `intercom-id-[app_id]`
    + `intercom-session-[app_id]`
    + `intercom-device-id-[app_id]`
+ google analytics
    ```
    _ga_WCZ03SZFCQ=GS1.1.1733858262.1.1.1733858267.55.0.0; 
    _ga=GA1.1.1448956542.1733858263; 
    _gid=GA1.2.397560962.1733858263; 
    _gat=1
    ```