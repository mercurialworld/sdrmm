package drm

import (
	"rustlang.pocha.moe/sdrmm/utils"
)

type MapData struct {
	BsrKey string
	Hash   string
	User   string // requester

	// title, subtitle, artist, mapper
	Title          string
	CensorTitle    bool
	SubTitle       string
	CensorSubTitle bool
	Artist         string
	CensorArtist   bool
	Mapper         string
	CensorMapper   bool

	// other map metadata
	Duration   int
	Votes      []int // upvotes, downvotes
	Rating     float64
	UploadTime utils.UnixTime // hopefully this works
	LastUpdate utils.UnixTime
	Cover      string
	Automapped bool

	// ranking
	ScoreSaberRanked bool
	BeatLeaderRanked bool
	Curated          bool
	CuratorName      string

	// in-game stuff
	Playlists  []string
	VoteStatus int

	// mod requirements
	UsesChroma            bool
	UsesCinema            bool
	UsesMappingExtensions bool
	UsesNoodleExtensions  bool
	UsesVivify            bool

	// caching
	DataIsFromLocalMap   bool
	DataIsFromLocalCache bool
	DataIsFromBeatSaver  bool

	// difficulties
	Diffs []MapDifficultyData
}

type MapDifficultyData struct {
	Difficulty      string
	Characteristic  string
	NoteJumpSpeed   float64
	NotesPerSecond  float64
	MapMods         IMapMods
	ScoreSaberStars float64
	BeatLeaderStars float64
}

type IMapMods struct {
	Chroma            bool
	Cinema            bool
	MappingExtensions bool
	NoodleExtensions  bool
	Vivify            bool
}

type QueuePositionData struct {
	Spot      int
	QueueItem MapData
}

type SessionHistoryItem struct {
	Timestamp   utils.UnixTime
	HistoryItem MapData
}

type Message struct {
	Message string
}
