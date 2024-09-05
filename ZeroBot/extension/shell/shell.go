package shell

import zero "ZeroBot"

func Parse(s string) []string {
	return zero.ParseShell(s)
}
