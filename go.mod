module github.com/dorfire/heavenly

go 1.19

require (
	github.com/deckarep/golang-set/v2 v2.1.0
	github.com/earthly/earthly v0.7.8
	github.com/earthly/earthly/ast v0.0.1
	github.com/go-git/go-git/v5 v5.7.0
	github.com/google/go-cmp v0.5.9
	github.com/samber/lo v1.38.1
	github.com/schollz/progressbar/v3 v3.13.1
	github.com/stretchr/testify v1.9.0
	github.com/tufin/asciitree v0.0.0-20210127111056-bf70173ef677
	github.com/urfave/cli/v2 v2.25.6
	golang.org/x/exp v0.0.0-20230224173230-c95f2b4c22f2
	golang.org/x/mod v0.10.0
)

require (
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230518184743-7afd39499903 // indirect
	github.com/acomagu/bufpipe v1.0.4 // indirect
	github.com/antlr/antlr4 v0.0.0-20200225173536-225249fdaef5 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.4.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/skeema/knownhosts v1.1.1 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/term v0.8.0 // indirect
	golang.org/x/tools v0.8.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Averts an error in GoLand
replace github.com/earthly/earthly/util/deltautil v0.0.0-00010101000000-000000000000 => github.com/earthly/earthly/util/deltautil v0.0.0-20230616170205-250e67318255
