package gin_plus

import (
	"os"
)

func getDefaultBanner() string {
	return "\033[38;2;60;120;255m   ______    _               ____     __                \n" +
		"\033[38;2;80;140;255m  / ____/   (_)   ____      / __ \\   / /  __  __   _____\n" +
		"\033[38;2;100;160;255m / / __    / /   / __ \\    / /_/ /  / /  / / / /  / ___/\n" +
		"\033[38;2;120;180;255m/ /_/ /   / /   / / / /   / ____/  / /  / /_/ /  (__  ) \n" +
		"\033[38;2;140;200;255m\\____/   /_/   /_/ /_/   /_/      /_/   \\____/  /____/   " +
		"\033[1;97m" + "(v4)" + "\033[0m" + "\n\n"
}

func printBanner(customBanner string) {
	if customBanner != "" {
		_, _ = os.Stderr.WriteString(customBanner + "\n")
	} else {
		_, _ = os.Stderr.WriteString(getDefaultBanner())
	}
}
