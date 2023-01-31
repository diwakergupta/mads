# wads
A tool to calculate weekly active developers based on Github activity.

This is **ALPHA** software, use at your own risk.

## Build

```bash
$ go get github.com/diwakergupta/wads
```

If that doesn't work, check out the repo and build locally `go build`

## Usage

* Find the TOML from EC's [crypto-ecosystem repo](https://github.com/electric-capital/crypto-ecosystems) that you want to process (e.g. [this one for Stacks](https://raw.githubusercontent.com/electric-capital/crypto-ecosystems/master/data/ecosystems/s/stacks.toml))
* Convert the TOML to JSON -- mostly because Go stdlib has built in support for JSON but not TOML. I used [this online convertor](https://pseitz.github.io/toml-to-json-online-converter/).
* Obtain a Github API token
* Run `wads --token <token> --file <file.json>` and wait


## Known limitations

This is a very hacky proof-of-concept. Lots of limitations and room for improvement:

* No tests
* Basic error handling. Does a naive, recursive retry if repo stats aren't yet available
* Doesn't discern commits from bots
* Doesn't handle "sub systems", therefore almost certainly undercounts
* Relies on Github repo stats being fresh & accurate. Directly obtaining data from a `git clone` will likely be more accurate
