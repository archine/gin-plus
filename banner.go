package gin_plus

// PrintBanner prints the application banner to the console.
var sysBanner = `
   ______    _               ____     __                
  / ____/   (_)   ____      / __ \   / /  __  __   _____
 / / __    / /   / __ \    / /_/ /  / /  / / / /  / ___/
/ /_/ /   / /   / / / /   / ____/  / /  / /_/ /  (__  ) 
\____/   /_/   /_/ /_/   /_/      /_/   \____/  /____/   (v4.0.6)
`

func printBanner() {
	if sysBanner != "" {
		println(sysBanner)
		sysBanner = ""
	}
}
