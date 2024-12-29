# gopaper
An Instapaper TUI client written in Go.

## Usage

Create a `.env` file at the root of the repo.

```bash
IP_USER=my-instapaper@email.com
IP_PASSWORD=mY1nstap@perP@ssw0rd
IP_API=https://www.instapaper.com/api
IP_API_VERSION=1.1
IP_OAUTH_CONSUMER_ID=xxxxxxx
IP_OAUTH_CONSUMER_SECRET=yyyyyy
```

Install and run the app.

```bash
$ go install .
$ gopaper
```

or

```bash
$ go run main.go
```

