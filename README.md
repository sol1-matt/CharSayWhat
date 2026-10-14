# CharSayWhat
This is a simple application that tells you what characters are in a string.

I asked AI to build this because I got annoyed trying to decypher the characters in passwords.

## Usage

```
charsaywhat                     # GUI
charsaywhat "P@ssw0rd"          # CLI, from arguments
echo "P@ssw0rd" | charsaywhat   # CLI, from a pipe
charsaywhat -s "P@ssw0rd"       # short type names by example character
charsaywhat -c "P@ssw0rd"       # short type names by two-letter code
charsaywhat -- -s               # analyse text that looks like a switch
```

| Type        | default       | `-s`        | `-c` |
|-------------|---------------|-------------|------|
| lowercase   | `lowercase`   | `a`         | `lc` |
| uppercase   | `uppercase`   | `A`         | `uc` |
| digit       | `digit`       | `9`         | `di` |
| space       | `space`       | ` ` (blank) | `sp` |
| punctuation | `punctuation` | `.`         | `pu` |
| symbol      | `symbol`      | `$`         | `sy` |
| newline     | `newline`     | `\n`        | `nl` |
| tab         | `tab`         | `\t`        | `tb` |
| other       | `other`       | `?`         | `ot` |

## Building

Binaries are written to `bin/`. A prebuilt Windows binary is committed at
`bin/charsaywhat.exe`; the Linux binary is not.

**Windows** (PowerShell):

```powershell
.\build.ps1          # -> bin\charsaywhat.exe
.\build.ps1 -Icon    # also regenerate the icon resource from charsaywhat.png
```

**Linux**:

```bash
./build.sh           # -> bin/charsaywhat
./build.sh windows   # cross-compile -> bin/charsaywhat.exe
./build.sh all       # both
./build.sh install   # build, then install binary + icon + menu launcher into ~/.local
```

Linux binaries can't carry an icon the way a Windows `.exe` does. The window
sets its own taskbar icon; `./build.sh install` adds the menu/launcher icon,
and in Caja/Nautilus the built binary is tagged with a custom file icon.

The Linux GUI uses [Gio](https://gioui.org), which needs cgo and some system
headers — see the comment at the top of `build.sh`.

## Tests

```
go test ./...
```
