package app

// PrintBanner prints the application banner to the console.
var sysBanner = `
   ______    _               ____     __                
  / ____/   (_)   ____      / __ \   / /  __  __   _____
 / / __    / /   / __ \    / /_/ /  / /  / / / /  / ___/
/ /_/ /   / /   / / / /   / ____/  / /  / /_/ /  (__  ) 
\____/   /_/   /_/ /_/   /_/      /_/   \____/  /____/   (v4.0.0)
`

func printBanner() {
   if sysBanner != "" {
      println(sysBanner)
   }
   sysBanner = ""
}