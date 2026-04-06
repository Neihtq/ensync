package sourceprovider

import (
	"io"
	"os"
	"path/filepath"
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
	_, err := provider.GetSource("non_existent_file.mp3")
	if err == nil {
		t.Errorf("Expected error when file is not found")
	}
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
	_, err = provider.GetSource(tmpFile.Name())
	if err == nil {
		t.Errorf("Expected error when file is invalid mp3")
	}
}

func TestMockSourceProvider_GetSource(t *testing.T) {
	provider := &MockSourceProvider{}

	filePath := "test_content"
	decoder, err := provider.GetSource(filePath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

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

func TestAudioProvider_SearchSong(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_audio_search")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.WriteFile(filepath.Join(tempDir, "oasis.mp3"), []byte("mock"), 0644)
	os.WriteFile(filepath.Join(tempDir, "blur.mp3"), []byte("mock"), 0644)

	provider := NewAudioProvider(tempDir)
	results := provider.SearchSong("oasis")

	if len(results) != 1 {
		t.Fatalf("Expected 1 result for 'oasis', got %d", len(results))
	}

	if results[0].Title != "oasis.mp3" {
		t.Errorf("Expected title 'oasis.mp3', got %s", results[0].Title)
	}

	// Test search without extension
	results = provider.SearchSong("blur")
	if len(results) != 1 {
		t.Fatalf("Expected 1 result for 'blur', got %d", len(results))
	}
	if results[0].Title != "blur.mp3" {
		t.Errorf("Expected title 'blur.mp3', got %s", results[0].Title)
	}
}

func TestMockSourceProvider_ListSongs(t *testing.T) {
	provider := &MockSourceProvider{}
	songs := provider.ListSongs()

	if len(songs) != 2 || songs[0] != "mock1" || songs[1] != "mock2" {
		t.Errorf("Expected mock songs ['mock1', 'mock2'], got %v", songs)
	}
}

func TestFloatToInt16_Clamping(t *testing.T) {
	if v := floatToInt16(0.0); v != 0 {
		t.Errorf("expected 0, got %d", v)
	}
	if v := floatToInt16(1.0); v != 32767 {
		t.Errorf("expected 32767, got %d", v)
	}
	if v := floatToInt16(-1.0); v != -32767 {
		t.Errorf("expected -32767, got %d", v)
	}
	if v := floatToInt16(2.5); v != 32767 {
		t.Errorf("expected clamped 32767, got %d", v)
	}
	if v := floatToInt16(-3.0); v != -32767 {
		t.Errorf("expected clamped -32767, got %d", v)
	}
}

func TestDecoder_Close(t *testing.T) {
	provider := &MockSourceProvider{}
	decoder, err := provider.GetSource("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = decoder.Close()
	if err != nil {
		t.Errorf("unexpected error on Close: %v", err)
	}
}

func TestDecoder_Read_EOF(t *testing.T) {
	provider := &MockSourceProvider{}
	decoder, _ := provider.GetSource("test")

	buf := make([]byte, 4096)
	totalRead := 0
	for {
		n, err := decoder.Read(buf)
		totalRead += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if totalRead == 0 {
		t.Error("expected to read some bytes before EOF")
	}
}

func TestAudioProvider_GetSource_Success(t *testing.T) {
	// Try to use a real mp3 from assets if it exists
	assetsDir := "../../../assets"
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		t.Skip("assets directory not found, skipping real mp3 test")
	}

	provider := NewAudioProvider(assetsDir)
	songs := provider.ListSongs()
	if len(songs) == 0 {
		t.Skip("no mp3 files in assets, skipping real mp3 test")
	}

	decoder, err := provider.GetSource(songs[0])
	if err != nil {
		t.Fatalf("Failed to get source for %s: %v", songs[0], err)
	}
	defer decoder.Close()

	if decoder.Streamer == nil {
		t.Fatal("expected non-nil streamer")
	}

	// Try reading a bit
	buf := make([]byte, 1024)
	n, err := decoder.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("Unexpected error reading: %v", err)
	}
	if n == 0 && err != io.EOF {
		t.Errorf("Expected to read some bytes")
	}
}
