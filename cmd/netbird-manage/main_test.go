package main

import "testing"

func TestConnectReturnsFlagParseError(t *testing.T) {
	if err := handleConnectCommand([]string{"connect", "--not-a-real-flag"}); err == nil {
		t.Fatal("invalid connect flag returned nil error")
	}
}
