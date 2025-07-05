package database

import "testing"

func Test_hashVideoUrl(t *testing.T) {
	result := hashVideoUrl("https://example.com/my_demo_file.mp3")
	result2 := hashVideoUrl("https://example.com/my_demo_file'.mp3")

	//the that result is a valid UUIDv4
	if result == "" {
		t.Errorf("hashVideoUrl() returned empty string")
	}

	if len(result) != 36 {
		t.Errorf("hashVideoUrl() returned string of length %d, want 36", len(result))
	}

	// Basic UUIDv4 format check: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	if result[14] != '4' {
		t.Errorf("hashVideoUrl() returned string with wrong version: %s", result)
	}
	if result[19] != '8' && result[19] != '9' && result[19] != 'a' && result[19] != 'b' {
		t.Errorf("hashVideoUrl() returned string with wrong variant: %s", result)
	}

	if result2 == result {
		t.Errorf("hashVideoUrl() returned same UUID for different input: %s", result)
	}
}
