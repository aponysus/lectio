package model

type SourceMedium string

const (
	SourceMediumBook         SourceMedium = "BOOK"
	SourceMediumEssay        SourceMedium = "ESSAY"
	SourceMediumArticle      SourceMedium = "ARTICLE"
	SourceMediumPaper        SourceMedium = "PAPER"
	SourceMediumScripture    SourceMedium = "SCRIPTURE"
	SourceMediumLecture      SourceMedium = "LECTURE"
	SourceMediumPodcast      SourceMedium = "PODCAST"
	SourceMediumFilm         SourceMedium = "FILM"
	SourceMediumTV           SourceMedium = "TV"
	SourceMediumConversation SourceMedium = "CONVERSATION"
	SourceMediumOther        SourceMedium = "OTHER"
)

var SourceMedia = []SourceMedium{
	SourceMediumBook,
	SourceMediumEssay,
	SourceMediumArticle,
	SourceMediumPaper,
	SourceMediumScripture,
	SourceMediumLecture,
	SourceMediumPodcast,
	SourceMediumFilm,
	SourceMediumTV,
	SourceMediumConversation,
	SourceMediumOther,
}

const (
	SourceSortRecent = "recent"
	SourceSortTitle  = "title"
)

type Source struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	Medium           string  `json:"medium"`
	Creator          *string `json:"creator,omitempty"`
	Year             *int    `json:"year,omitempty"`
	OriginalLanguage *string `json:"original_language,omitempty"`
	CultureOrContext *string `json:"culture_or_context,omitempty"`
	Notes            *string `json:"notes,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	ArchivedAt       *string `json:"archived_at,omitempty"`
}

type SourceInput struct {
	Title            string
	Medium           string
	Creator          string
	Year             *int
	OriginalLanguage string
	CultureOrContext string
	Notes            string
}

type SourceFilters struct {
	Query            string
	Medium           string
	OriginalLanguage string
	Sort             string
	Limit            int
	IncludeArchived  bool
}

func IsValidSourceMedium(value string) bool {
	for _, medium := range SourceMedia {
		if string(medium) == value {
			return true
		}
	}
	return false
}
