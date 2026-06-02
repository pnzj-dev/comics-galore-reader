module github.com/pnzj-dev/comics-galore-reader

go 1.25.1

require (
	github.com/nwaples/rardecode/v2 v2.0.1
	golang.org/x/image v0.40.0
)

require github.com/a-h/templ v0.3.1020

retract v0.1.0 // Module path was wrong (cg-reader-wasm); use v0.1.1+.
