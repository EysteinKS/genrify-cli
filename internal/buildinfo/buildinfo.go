package buildinfo

// Keep these in a tiny package so all other packages can reference
// consistent app/version/user-agent without import cycles.
//
// Override Version at build time:
//   go build -ldflags "-X genrify/internal/buildinfo.Version=0.1.0" ./cmd/genrify

const AppName = "genrify"

var Version = "dev"

var UserAgent = AppName + "/" + Version
