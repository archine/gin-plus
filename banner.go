package gin_plus

import (
	"fmt"
	"os"
)

// PrintBanner prints the application banner to the console.
var sysBanner = `
   ______    _               ____     __                
  / ____/   (_)   ____      / __ \   / /  __  __   _____
 / / __    / /   / __ \    / /_/ /  / /  / / / /  / ___/
/ /_/ /   / /   / / / /   / ____/  / /  / /_/ /  (__  ) 
\____/   /_/   /_/ /_/   /_/      /_/   \____/  /____/   (v4.1.9)
`

func printBanner() {
	if sysBanner != "" {
		_, _ = fmt.Fprint(os.Stderr, sysBanner)
		sysBanner = ""
	}
}
