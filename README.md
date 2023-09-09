# rage

[_Because we all get a little pissed sometimes..._](https://twitter.com/trevorsuarez/status/507181061861015553)

#### What?

This app has a LONG history, and unfortunately it's referenceable past has experienced [link rot](https://en.wikipedia.org/wiki/Link_rot). In short, a co-worker of mine from many years ago had a silly bash alias that worked by pinging a service like this, and then it died. Soooo, I recreated it... and then it sat for 10+ years untouched. And yea, amazingly it ran without issues all that time.

Recently, I ([Trevor Suarez (Rican7)](https://trevorsuarez.com/)) started some maintenance on some old projects and servers, and I found this app just sitting, still running all this time, but on a VERY old version of PHP, with it's version-controlled source having never been pushed to a remote... So, I pushed the source to my GitHub, containerized the app, and now it's running on modern serverless fully-managed compute. Yay!

And, well, then I decided to re-write it in Go! Why? Well, because honestly the old PHP source would have taken a bit of effort just to update to be properly compatible (and secure) with modern PHP versions, and because PHP isn't super well-suited for these kinds of serverless workloads. Oh, yea, and just because I love Go. :)

Anyway, this is silly. Enjoy!

PS: Alias this in your shell environment for a good time, like this:

```shell
alias fuck="curl -Ls https://rage.metroserve.me/?format=plain"
```


## Install

> [!IMPORTANT]
> This service requires a Redis instance, of which it connects via information provided in environment variables.

If you have a working Go environment, you can install via `go install`:

```shell
go install github.com/Rican7/rage@latest
```

... Otherwise, if you have Docker, you should be able to just build it with

`docker build`:

```shell
docker build -t 'rage-or-whatever-you-wanna-call-it' .
```


## Configuration

This app is predominantly configured via environment variables.

To see what environment variables are available to configure, see the
`.env.example` file.
