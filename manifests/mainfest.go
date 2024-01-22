package manifests

import "embed"

// FS embeds the manifests
//
//go:embed virtualcluster/*
//go:embed runnergroup/*
var FS embed.FS
