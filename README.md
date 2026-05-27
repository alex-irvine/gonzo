# gonzofk

A fork of [control-theory/gonzo](https://github.com/control-theory/gonzo) — a real-time log analysis TUI inspired by k9s.

The binary is renamed **`gonzofk`** so it installs alongside upstream `gonzo` without clobbering it. Everything works the same; just substitute `gonzofk` for `gonzo`. For full feature docs, AI/OTLP/custom-format guides, and integration examples, see the [upstream project](https://github.com/control-theory/gonzo) and [its docs](https://docs.controltheory.com/).

## Installation

### Prebuilt release (recommended)

Releases are published on tag push (`v*`). Download with the [GitHub CLI](https://cli.github.com):

```bash
# Linux (amd64)
mkdir -p ~/.local/bin
gh release download --repo alex-irvine/gonzo --pattern 'gonzofk-linux-amd64' --output ~/.local/bin/gonzofk --clobber
chmod +x ~/.local/bin/gonzofk

# macOS (Apple Silicon)
gh release download --repo alex-irvine/gonzo --pattern 'gonzofk-darwin-arm64' --output ~/.local/bin/gonzofk --clobber
chmod +x ~/.local/bin/gonzofk
```

Make sure `~/.local/bin` is on your `PATH`.

### Build from source

```bash
git clone https://github.com/alex-irvine/gonzo.git
cd gonzo
make build            # outputs ./build/gonzofk
make install          # installs gonzofk to $GOPATH/bin
```

Or with Go directly:

```bash
go build -o gonzofk ./cmd/gonzofk
# NOTE: `go run main.go` will NOT work — the package spans multiple files.
go run ./cmd/gonzofk
```

### Nix (flake)

```bash
nix run github:alex-irvine/gonzo
```

## Usage

```bash
# Pipe from stdin
tail -f /var/log/app.log | gonzofk
cat application.log | gonzofk
kubectl logs -f deployment/my-app | gonzofk
docker logs -f my-container 2>&1 | gonzofk
journalctl -f | gonzofk

# Read files directly
gonzofk -f application.log
gonzofk -f app.log -f error.log              # multiple files
gonzofk -f "/var/log/*.log"                  # glob
gonzofk -f /var/log/app.log --follow         # follow, like tail -f

# Stream from Kubernetes
gonzofk --k8s-enabled=true --k8s-namespaces=default
gonzofk --k8s-enabled=true --k8s-selector="app=my-app"

# With AI analysis (requires API key)
export OPENAI_API_KEY=sk-your-key-here
gonzofk -f application.log --ai-model="gpt-4"
```

Run `gonzofk --help` for the full flag list.

### Key bindings

| Key | Action |
| --- | --- |
| `↑`/`↓` or `k`/`j` | Move selection / scroll |
| `Tab` / `Shift+Tab` | Switch panels |
| `Space` | Pause / unpause dashboard |
| `/` | Filter mode (regex) |
| `s` | Search & highlight |
| `f` | Fullscreen log viewer |
| `Enter` | Log details |
| `v` | Visual selection (vim-style) |
| `y` | Yank selection to clipboard |
| `?` | Help |
| `q` / `Ctrl+C` | Quit |

## License

MIT — see [LICENSE](LICENSE). Original work by [ControlTheory](https://controltheory.com) and the Gonzo community.
