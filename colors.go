// colors.go
package main

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

var (
	// colorsEnabled is set based on terminal detection
	colorsEnabled = false
)

func init() {
	// Enable colors if output is a TTY
	colorsEnabled = isTTY()
}

// isTTY checks if the output is going to a terminal
func isTTY() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// colorize wraps text with ANSI color codes if colors are enabled
func colorize(color, text string) string {
	if !colorsEnabled {
		return text
	}
	return color + text + ColorReset
}

// Color helper functions for common use cases
func red(text string) string {
	return colorize(ColorRed, text)
}

func green(text string) string {
	return colorize(ColorGreen, text)
}

func yellow(text string) string {
	return colorize(ColorYellow, text)
}

func blue(text string) string {
	return colorize(ColorBlue, text)
}

func cyan(text string) string {
	return colorize(ColorCyan, text)
}

func purple(text string) string {
	return colorize(ColorPurple, text)
}

func bold(text string) string {
	return colorize(ColorBold, text)
}

func dim(text string) string {
	return colorize(ColorDim, text)
}

// Semantic color functions
func success(text string) string {
	return green("✓ " + text)
}

func failure(text string) string {
	return red("✗ " + text)
}

func warning(text string) string {
	return yellow("⚠️  " + text)
}

func info(text string) string {
	return cyan("ℹ " + text)
}

func header(text string) string {
	return bold(cyan(text))
}

// Status-specific formatters
func statusOnline(text string) string {
	return green(text)
}

func statusOffline(text string) string {
	return red(text)
}

func statusEnabled(enabled bool) string {
	if enabled {
		return green("Enabled")
	}
	return red("Disabled")
}

func statusConnected(connected bool) string {
	if connected {
		return green("Online")
	}
	return red("Offline")
}

// ID formatter - dims UUIDs to reduce visual noise
func formatID(id string) string {
	return dim(id)
}

// Colorized printf wrappers
func printSuccess(format string, a ...interface{}) {
	fmt.Printf(success(fmt.Sprintf(format, a...)) + "\n")
}

func printFailure(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, failure(fmt.Sprintf(format, a...))+"\n")
}

func printWarning(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, warning(fmt.Sprintf(format, a...))+"\n")
}

func printInfo(format string, a ...interface{}) {
	fmt.Printf(info(fmt.Sprintf(format, a...)) + "\n")
}

func printHeader(format string, a ...interface{}) {
	fmt.Printf(header(fmt.Sprintf(format, a...)) + "\n")
}
