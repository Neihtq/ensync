package navidrome

type ResponseWrapper struct {
	SubsonicResponse SubsonicResponse `json:"subsonic-response"`
}

type SubsonicResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Error   *APIError `json:"error,omitempty"`

	SearchResult3 *SearchResult3 `json:"searchResult3,omitempty"`
	Song          *Song          `json:"song,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SearchResult3 struct {
	Artist []Artist `json:"artist,omitempty"`
	Album  []Album  `json:"album,omitempty"`
	Song   []Song   `json:"song,omitempty"`
}

type Artist struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CoverArt   string `json:"coverArt,omitempty"`
	AlbumCount int    `json:"albumCount,omitempty"`
}

type Album struct {
	ID        string `json:"id"`
	Title     string `json:"name"`
	Artist    string `json:"artist"`
	ArtistID  string `json:"artistId"`
	CoverArt  string `json:"coverArt,omitempty"`
	SongCount int    `json:"songCount"`
	Duration  int    `json:"duration,omitempty"`
	Year      int    `json:"year,omitempty"`
}

type Song struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Album       string `json:"album"`
	AlbumID     string `json:"albumId"`
	Artist      string `json:"artist"`
	ArtistID    string `json:"artistId"`
	Track       int    `json:"track,omitempty"`
	Year        int    `json:"year,omitempty"`
	Genre       string `json:"genre,omitempty"`
	CoverArt    string `json:"coverArt,omitempty"`
	Size        int64  `json:"size,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Suffix      string `json:"suffix,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	BitRate     int    `json:"bitRate,omitempty"`
	Path        string `json:"path,omitempty"`
}
