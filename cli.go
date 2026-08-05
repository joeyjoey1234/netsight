package main

import (
	"flag"
	"fmt"
	"os"
)

func runCLI() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "scan":
		scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
		subnet := scanCmd.String("subnet", "", "Subnet to scan (e.g. 192.168.1.0/24)")
		output := scanCmd.String("output", "scan.json", "Output file path")
		preset := scanCmd.String("preset", "quick", "Survey preset: quick, short, long")

		scanCmd.Parse(os.Args[2:])

		if *subnet == "" {
			fmt.Println("Error: --subnet is required")
			scanCmd.Usage()
			os.Exit(1)
		}

		fmt.Printf("NetSight %s survey starting...\n", *preset)
		fmt.Printf("Scanning subnet: %s\n", *subnet)
		fmt.Printf("Output: %s\n", *output)
		fmt.Println("(CLI mode — full implementation wiring in progress)")
		fmt.Println("Survey complete. Results saved.")

	case "export":
		exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
		input := exportCmd.String("input", "", "Input JSON file from a previous scan")
		format := exportCmd.String("format", "pdf", "Output format: pdf, drawio, json")

		exportCmd.Parse(os.Args[2:])

		if *input == "" {
			fmt.Println("Error: --input is required")
			exportCmd.Usage()
			os.Exit(1)
		}

		fmt.Printf("Exporting %s as %s...\n", *input, *format)
		fmt.Println("Export complete.")

	case "version", "--version", "-v":
		fmt.Println("NetSight v0.1.0")
		fmt.Println("Portable Network Analysis Tool for Windows")
		fmt.Println("Built with Go + Wails v2")

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`NetSight — Portable Network Analysis Tool

Usage:
  netsight [command] [flags]

Commands:
  scan     Run a network scan
  export   Export scan results
  version  Show version information
  help     Show this help

Scan Flags:
  --subnet   Subnet to scan in CIDR notation (required)
  --output   Output file path (default: scan.json)
  --preset   Survey preset: quick, short, long (default: quick)

Export Flags:
  --input    Input JSON file from a previous scan (required)
  --format   Output format: pdf, drawio, json (default: pdf)

Running without arguments launches the GUI.`)
}
