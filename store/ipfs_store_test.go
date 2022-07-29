package store

import (
	"testing"

	"golang.org/x/net/context"
)

func TestDecrypt(t *testing.T) {

	store := NewIpfsStore("10.1.1.29:5001")
	_, err := store.GetFile(context.Background(), map[string]string{"hash": "QmatXjoufeJbAc9Aw2TDpeDKHMfPLgwzBKdC51ovEZxZrX"})
	if err != nil {
		t.Fatal("failed to get file", err)
	}
}
