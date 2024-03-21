# AuraGem Server(s)

Search Engine, Star Wars Database, Weather, Music Storage, etc.

## Setting It Up

Set up by copying `config/config.go.example` to `config/config.go` and setting up your DB info. Create databases in Firebird for music and search, and set the locations in the config.go file. Then, go into `gemini/gemini.go` and change the hostnames and certificates for each server. Run `go build .` to build the executable.

To create the database tables, run `auragem_sis migrate search` and `auragem_sis migrate music`. For other databases, run the same command using the database's name.

Lastly, to handle Full-Text Search for the Search Engine, install the udr lucene plugin for firebird and run the queries in `migration/migrations/fts.sql`. Start the server by running `auragem_sis`.

## License Info
This capsule is currently licensed as BSD-3-Clause. Below is a list of libraries that are used and their licenses.

* [clseibold/smallnetinformationservices](https://gitlab.com/clseibold/smallnetinformationservices) - BSD-2 Clause - My Server Software

* [nakagami/firebirdsql](https://github.com/nakagami/firebirdsql) - [MIT](https://opensource.org/licenses/MIT)
* [google/go-github](https://github.com/google/go-github/) - Used for GitHub proxy
* [dhowden/tag](https://github.com/dhowden/tag) - [BSD-3 Clause](https://opensource.org/licenses/BSD-2-Clause)
* [juju/ratelimit](https://github.com/juju/ratelimit) - Used for rate-limiting music streaming to a specific kbps
* [kkdai/youtube](https://github.com/kkdai/youtube) - MIT - Used for YouTube proxy
* [krayzpipes/cronticker](https://github.com/krayzpipes/cronticker) - MIT - Used for gemini live chat, to send a daily system message and clear the chat history.

* [golang.org/x/time](https://golang.org/x/time) - BSD-3 Clause - Used for time rate limiting
* [golang.org/x/net](https://golang.org/x/net) - BSD-3 Clause - Networking stuff
* [golang.org/x/text](https://golang.org/x/text) - BSD-3 Clause - Used for text processing
* [google.golang.org/api](https://google.golang.org/api) - BSD-3 Clause - Used for YouTube proxy
* [golang.org/x/oauth2](https://golang.org/x/oauth2) - BSD-3 Clause - Oauth Client used for GitHub proxy.

* [rs/zerolog](https://github.com/rs/zerolog) - MIT - Logging
* [spf13/cobra](https://github.com/spf13/cobra) - Apache 2.0 - Command line stuff

* Google Golang Standard Library - [BSD-3 Clause](https://opensource.org/licenses/BSD-3-Clause)
