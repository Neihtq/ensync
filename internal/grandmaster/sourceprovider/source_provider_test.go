package sourceprovider

import (
	"io"
	"testing"
)

const testDir = "./some/test"

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

func TestMockSourceProvider_ListSongs(t *testing.T) {
	provider := &MockSourceProvider{}
	songs := provider.ListSongs()

	if len(songs) != 0 {
		t.Errorf("Expected no mock songs, got %v", songs)
	}
}

func TestMockSourceProvider_GetTitle(t *testing.T) {
	provider := &MockSourceProvider{}
	trackID := "mock_track"
	if provider.GetTitle(trackID) != trackID {
		t.Errorf("expected %s, got %s", trackID, provider.GetTitle(trackID))
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
