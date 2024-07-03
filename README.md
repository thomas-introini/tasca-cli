Tasca
-----

Tasca is a TUI [Pocket](https://getpocket.com/) client written in Go.

## How to build

```bash
make build
```

## How to run

Right now the only supported way of running is to pass the _Pocket_ [consumer key](http://getpocket.com/developer/apps/new) as an environment variable and then run the binary like this:

```bash
POCKET_CONSUMER_KEY=<your_consumer_key> ./tasca
```
