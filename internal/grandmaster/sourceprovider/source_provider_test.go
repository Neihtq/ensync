package sourceprovider

import (
	"io"
	"os"
	"testing"
)

const testDir = "./some/test"

func TestAudioProvider_New(t *testing.T) {
	provider := NewAudioProvider(testDir)
	if provider == nil {
		t.Fatalf("Expected non-nil provider")
	}
}

func TestAudioProvider_GetSource_FileNotFound(t *testing.T) {
	provider := NewAudioProvider(testDir)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when file is not found")
		}
	}()
	provider.GetSource("non_existent_file.mp3")
}

func TestAudioProvider_GetSource_InvalidMp3(t *testing.T) {
	// Create a temporary invalid mp3 file
	tmpFile, err := os.CreateTemp("", "invalid*.mp3")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("this is not an mp3 file")
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	provider := NewAudioProvider(testDir)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when file is invalid mp3")
		}
	}()
	provider.GetSource(tmpFile.Name())
}

func TestMockSourceProvider_GetSource(t *testing.T) {
	provider := &MockSourceProvider{}

	filePath := "test_content"
	decoder := provider.GetSource(filePath)

	if decoder == nil {
		t.Fatalf("Expected non-nil decoder")
	}

	if decoder.SampleRate != 48000 {
		t.Errorf("Expected sample rate 48000, got %d", decoder.SampleRate)
	}

	if decoder.Channels != 2 {
		t.Errorf("Expected 2 channels, got %d", decoder.Channels)
	}

	buffer := make([]byte, 1024)
	n, err := decoder.Read(buffer)

	if err != nil && err != io.EOF {
		t.Errorf("Unexpected error reading from decoder: %v", err)
	}

	if n == 0 {
		t.Errorf("Expected to read some bytes")
	}

	content := string(buffer[:n])
	if content != filePath {
		t.Errorf("Expected to read '%s', got '%s'", filePath, content)
	}

	err = decoder.Close()
	if err != nil {
		t.Errorf("Unexpected error closing decoder: %v", err)
	}
}

func TestAudioProvider_ListSongs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_audio_root")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Create(tempDir + "/song1.mp3")
	os.Create(tempDir + "/song2.mp3")
	os.Create(tempDir + "/not_a_song.txt")

	provider := NewAudioProvider(tempDir)
	songs := provider.ListSongs()

	if len(songs) != 2 {
		t.Errorf("Expected 2 songs, got %d. Songs: %v", len(songs), songs)
	}
}

func TestMockSourceProvider_ListSongs(t *testing.T) {
	provider := &MockSourceProvider{}
	songs := provider.ListSongs()

	if len(songs) != 2 || songs[0] != "mock1" || songs[1] != "mock2" {
		t.Errorf("Expected mock songs ['mock1', 'mock2'], got %v", songs)
	}
}
