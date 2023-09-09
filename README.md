# rage

## What?

This app has a LONG history, and unfortunately it's referenceable past has experienced link rot. In short, a co-worker of mine from many years ago had a silly bash alias that worked like this, and then it died. Soooo, I recreated it... and then it sat for 10+ years untouched. And yea, amazingly it ran without issues all that time.

Recently, I (Trevor Suarez (Rican7)) started some maintenance on some old projects and servers, and I found this app just sitting, still running all this time, but on a VERY old version of PHP, with it's version-controlled source having never been pushed to a remote... So, I pushed the source to my GitHub, containerized the app, and now it's running on modern serverless fully-managed compute. Yay!

And, well, then I decided to re-write it in Go! Why? Well, because honestly the old PHP source would have taken a bit of effort just to update to be properly compatible (and secure) with modern PHP versions, and because PHP isn't super well-suited for these kinds of serverless workloads. Oh, yea, and just because I love Go. :)

Anyway, this is silly. Enjoy!

PS: Alias this in your shell environment for a good time, like this:

    alias fuck="curl -Ls https://rage.metroserve.me/?format=plain"

