package main

import (
	_ "embed"
)

//go:embed static/bootstrap.min.css
var bootstrapCSS []byte

//go:embed static/bootstrap.bundle.min.js
var bootstrapJS []byte