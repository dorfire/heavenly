# heavenly

Earthly tools.

Motivated by some tools that exist in the Bazel ecosystem but are missing in the Earthly world:

- [buildifier](https://github.com/bazelbuild/buildtools/blob/master/buildifier/README.md)
- [Gazelle](https://github.com/bazelbuild/bazel-gazelle/)
- [bazel-diff](https://github.com/Tinder/bazel-diff)

## Usage

```
NAME:
   heavenly - manages Earthly from above

USAGE:
   heavenly [global options] command [command options] [arguments...]

DESCRIPTION:
   heavenly is a CLI tool that formats, lints and analyzes Earthly repos and the Earthfiles in them.

COMMANDS:
   format, fmt      format Earthfiles in the current repo according to a set of rules
   lint             lint the current repo according to a set of rules
   changed          analyze a given Earthly target and exit with 0 if it has any changed input files. exit with 1 otherwise.
   matrix           analyze a given Earthly target and output the BUILD commands within it that need rebuilding for a given git diff
   matrix-deps      analyze a given Earthly target and output the BUILD commands within it that need rebuilding for a given set of changed input files
   inspect, inputs  analyze a given Earthly target and show which source files it depends on
   gocopies         analyze a given Go package and print the COPY commands it needs in order to build
   help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --chdir value  
   --debug        (default: false)
   --help, -h     show help
```
