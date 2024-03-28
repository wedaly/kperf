package manifests

import "embed"

// FS embeds the manifests
//
//go:embed workload/*
//go:embed loadprofile/*
var FS embed.FS
