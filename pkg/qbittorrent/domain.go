package qbittorrent

type Torrent struct {
	AddedOn            int          `json:"added_on"`
	AmountLeft         int          `json:"amount_left"`
	AutoManaged        bool         `json:"auto_tmm"`
	Availability       float32      `json:"availability"`
	Category           string       `json:"category"`
	Completed          int          `json:"completed"`
	CompletionOn       int          `json:"completion_on"`
	DlLimit            int          `json:"dl_limit"`
	DlSpeed            int          `json:"dlspeed"`
	Downloaded         int          `json:"downloaded"`
	DownloadedSession  int          `json:"downloaded_session"`
	ETA                int          `json:"eta"`
	FirstLastPiecePrio bool         `json:"f_l_piece_prio"`
	ForceStart         bool         `json:"force_start"`
	Hash               string       `json:"hash"`
	LastActivity       int          `json:"last_activity"`
	MagnetURI          string       `json:"magnet_uri"`
	MaxRatio           float32      `json:"max_ratio"`
	MaxSeedingTime     int          `json:"max_seeding_time"`
	Name               string       `json:"name"`
	NumComplete        int          `json:"num_complete"`
	NumIncomplete      int          `json:"num_incomplete"`
	NumSeeds           int          `json:"num_seeds"`
	Priority           int          `json:"priority"`
	Progress           float32      `json:"progress"`
	Ratio              float32      `json:"ratio"`
	RatioLimit         float32      `json:"ratio_limit"`
	SavePath           string       `json:"save_path"`
	SeedingTimeLimit   int          `json:"seeding_time_limit"`
	SeenComplete       int          `json:"seen_complete"`
	SequentialDownload bool         `json:"seq_dl"`
	Size               int          `json:"size"`
	State              TorrentState `json:"state"`
	SuperSeeding       bool         `json:"super_seeding"`
	Tags               string       `json:"tags"`
	TimeActive         int          `json:"time_active"`
	TotalSize          int          `json:"total_size"`
	Tracker            string       `json:"tracker"`
	UpLimit            int          `json:"up_limit"`
	Uploaded           int          `json:"uploaded"`
	UploadedSession    int          `json:"uploaded_session"`
	UpSpeed            int          `json:"upspeed"`
}

type TorrentBasic struct {
	Category   string       `json:"category"`
	Downloaded int          `json:"downloaded"`
	Hash       string       `json:"hash"`
	Name       string       `json:"name"`
	Progress   float32      `json:"progress"`
	Ratio      float32      `json:"ratio"`
	Size       int          `json:"size"`
	State      TorrentState `json:"state"`
	Tags       string       `json:"tags"`
	Tracker    string       `json:"tracker"`
	Uploaded   int          `json:"uploaded"`
}

type TorrentTrackersResponse struct {
	Trackers []TorrentTracker `json:"trackers"`
}

type TorrentTracker struct {
	//Tier          uint   `json:"tier"` // can be both empty "" and int
	Url           string `json:"url"`
	Status        int    `json:"status"`
	NumPeers      int    `json:"num_peers"`
	NumSeeds      int    `json:"num_seeds"`
	NumLeechers   int    `json:"num_leechers"`
	NumDownloaded int    `json:"num_downloaded"`
	Message       string `json:"msg"`
}

type TorrentState string

const (
	// Some error occurred, applies to paused torrents
	TorrentStateError TorrentState = "error"

	// Torrent data files is missing
	TorrentStateMissingFiles TorrentState = "missingFiles"

	// Torrent is being seeded and data is being transferred
	TorrentStateUploading TorrentState = "uploading"

	// Torrent is paused and has finished downloading
	TorrentStatePausedUp TorrentState = "pausedUP"

	// Queuing is enabled and torrent is queued for upload
	TorrentStateQueuedUp TorrentState = "queuedUP"

	// Torrent is being seeded, but no connection were made
	TorrentStateStalledUp TorrentState = "stalledUP"

	// Torrent has finished downloading and is being checked
	TorrentStateCheckingUp TorrentState = "checkingUP"

	// Torrent is forced to uploading and ignore queue limit
	TorrentStateForcedUp TorrentState = "forcedUP"

	// Torrent is allocating disk space for download
	TorrentStateAllocating TorrentState = "allocating"

	// Torrent is being downloaded and data is being transferred
	TorrentStateDownloading TorrentState = "downloading"

	// Torrent has just started downloading and is fetching metadata
	TorrentStateMetaDl TorrentState = "metaDL"

	// Torrent is paused and has NOT finished downloading
	TorrentStatePausedDl TorrentState = "pausedDL"

	// Queuing is enabled and torrent is queued for download
	TorrentStateQueuedDl TorrentState = "queuedDL"

	// Torrent is being downloaded, but no connection were made
	TorrentStateStalledDl TorrentState = "stalledDL"

	// Same as checkingUP, but torrent has NOT finished downloading
	TorrentStateCheckingDl TorrentState = "checkingDL"

	// Torrent is forced to downloading to ignore queue limit
	TorrentStateForceDl TorrentState = "forceDL"

	// Checking resume data on qBt startup
	TorrentStateCheckingResumeData TorrentState = "checkingResumeData"

	// Torrent is moving to another location
	TorrentStateMoving TorrentState = "moving"

	// Unknown status
	TorrentStateUnknown TorrentState = "unknown"
)

// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)#get-torrent-list
type TorrentFilter string

const (
	// Torrent is paused
	TorrentFilterAll TorrentFilter = "all"

	// Torrent is active
	TorrentFilterActive TorrentFilter = "active"

	// Torrent is inactive
	TorrentFilterInactive TorrentFilter = "inactive"

	// Torrent is completed
	TorrentFilterCompleted TorrentFilter = "completed"

	// Torrent is resumed
	TorrentFilterResumed TorrentFilter = "resumed"

	// Torrent is paused
	TorrentFilterPaused TorrentFilter = "paused"

	// Torrent is seeding
	TorrentFilterSeeding TorrentFilter = "seeding"

	// Torrent is stalled
	TorrentFilterStalled TorrentFilter = "stalled"

	// Torrent is being seeded and data is being transferred
	TorrentFilterUploading TorrentFilter = "uploading"

	// Torrent is being seeded, but no connection were made
	TorrentFilterStalledUploading TorrentFilter = "stalled_uploading"

	// Torrent is being downloaded and data is being transferred
	TorrentFilterDownloading TorrentFilter = "downloading"

	// Torrent is being downloaded, but no connection were made
	TorrentFilterStalledDownloading TorrentFilter = "stalled_downloading"

	TorrentFilterErrored TorrentFilter = "errored"
)

type GetTorrentsRequest struct {
	Filter   string
	Category string
	Tags     string
	Hashes   string
	Json     bool
}
