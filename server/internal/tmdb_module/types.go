package tmdbmodule

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductionCompany struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LogoPath string `json:"logo_path"`
}

type PaginatedResult[T any] struct {
	Page         int `json:"page"`
	Results      []T `json:"results"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

type Media struct {
	ID        int    `json:"id"`
	MediaType string `json:"media_type"`
	Title     string `json:"title"`
	PosterURL string `json:"poster_url"`
}

type Movie struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	ReleaseDate   string  `json:"release_date"`
	VoteAverage   float32 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	Popularity    float32 `json:"popularity"`
	GenreIDs      []int   `json:"genre_ids"`
	Adult         bool    `json:"adult"`
}

type MovieDetail struct {
	Movie
	EnglishTitle        string              `json:"english_title,omitempty"`
	IMDbID              string              `json:"imdb_id,omitempty"`
	Runtime             int                 `json:"runtime"`
	Status              string              `json:"status"`
	Tagline             string              `json:"tagline"`
	Budget              int                 `json:"budget"`
	Revenue             int                 `json:"revenue"`
	Homepage            string              `json:"homepage"`
	Genres              []Genre             `json:"genres"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
}

type Cast struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
	Order       int    `json:"order"`
}

type Crew struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Job         string `json:"job"`
	Department  string `json:"department"`
	ProfilePath string `json:"profile_path"`
}

type Credits struct {
	Cast []Cast `json:"cast"`
	Crew []Crew `json:"crew"`
}

type Video struct {
	ID       string `json:"id"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	Site     string `json:"site"`
	Type     string `json:"type"`
	Official bool   `json:"official"`
}

type VideoResult struct {
	Results []Video `json:"results"`
}

type TVSeries struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	OriginalName string  `json:"original_name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	FirstAirDate string  `json:"first_air_date"`
	VoteAverage  float32 `json:"vote_average"`
	VoteCount    int     `json:"vote_count"`
	Popularity   float32 `json:"popularity"`
	GenreIDs     []int   `json:"genre_ids"`
	Adult        bool    `json:"adult"`
}

type TVSeriesDetail struct {
	TVSeries
	EnglishTitle        string              `json:"english_title,omitempty"`
	NumberOfSeasons     int                 `json:"number_of_seasons"`
	NumberOfEpisodes    int                 `json:"number_of_episodes"`
	Status              string              `json:"status"`
	Tagline             string              `json:"tagline"`
	Homepage            string              `json:"homepage"`
	Genres              []Genre             `json:"genres"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
}

type Person struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	ProfilePath  string  `json:"profile_path"`
	KnownForDept string  `json:"known_for_department"`
	Popularity   float32 `json:"popularity"`
}

type PersonDetail struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	Biography          string  `json:"biography"`
	Birthday           *string `json:"birthday"`
	Deathday           *string `json:"deathday"`
	PlaceOfBirth       *string `json:"place_of_birth"`
	ProfilePath        *string `json:"profile_path"`
	KnownForDepartment string  `json:"known_for_department"`
	Popularity         float32 `json:"popularity"`
	Gender             int     `json:"gender"`
	IMDBID             *string `json:"imdb_id"`
	Homepage           *string `json:"homepage"`
}

type PersonMovieCreditItem struct {
	Movie
	Character string `json:"character"`
	CreditID  string `json:"credit_id"`
}

type PersonMovieCredits struct {
	Cast []PersonMovieCreditItem `json:"cast"`
	Crew []Movie                 `json:"crew"`
}

type PersonTVCreditItem struct {
	TVSeries
	Character    string `json:"character"`
	CreditID     string `json:"credit_id"`
	EpisodeCount int    `json:"episode_count"`
}

type PersonTVCredits struct {
	Cast []PersonTVCreditItem `json:"cast"`
	Crew []TVSeries           `json:"crew"`
}
