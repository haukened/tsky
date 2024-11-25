package utils

var version = "dev"

func SetVersion(v string) {
	version = v
}

func UserAgent() string {
	return "tsky/" + version
}
