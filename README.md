# pr-url-scanner
This service scans PRs for URLs, verifies whether they are reachable, and reports the results in a comment.

## To build:
```
go get
go build -o main
```

## To run:
```
GITHUB_TOKEN=<your-token> ./main
```

## Dependencies:
mvdan.cc/xurls

## Todo items:
- verify github signature on webhook
- move config vars to env + defaults (webhook secret, github token, port)
- use versioned dependencies
- set max number of urls from env to prevent abuse
- add some comments to code
- add `go test`