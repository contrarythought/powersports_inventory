package main

import "testing"

func TestConnectDB(t *testing.T) {
	db, err := connectDB()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
}
