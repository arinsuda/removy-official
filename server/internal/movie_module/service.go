package movie_module

import (
	"context"
	"fmt"

	external "github.com/arinsuda/movie-hub/internal/external_rating_module"
	"github.com/arinsuda/movie-hub/internal/shared"
	tmdb "github.com/arinsuda/movie-hub/internal/tmdb_module"
)

type MovieService struct {
	ratingRepo shared.RatingStatsReader
	external   *external.Service

	providers      *providersCache
	fetchProviders func(kind string, id int, region string) (*tmdb.WatchProviders, error)
}

// SetExternalRatings enables IMDb and Rotten Tomatoes on detail responses.
func (s *MovieService) SetExternalRatings(ext *external.Service) { s.external = ext }

func NewMovieService(repo shared.RatingStatsReader) *MovieService {
	return &MovieService{ratingRepo: repo, providers: newProvidersCache(), fetchProviders: tmdb.GetWatchProviders}
}

func (s *MovieService) GetPopular(ctx context.Context, page int) (*tmdb.PaginatedResult[MovieDTO], error) {
	res, err := tmdb.GetPopular(page)
	if err != nil {
		return nil, err
	}
	s.prewarmMovies(res.Results)
	return s.enrichMovies(ctx, res)
}

func (s *MovieService) GetNowPlaying(ctx context.Context, page int, options tmdb.RequestOptions) (*tmdb.PaginatedResult[MovieDTO], error) {
	res, err := tmdb.GetNowPlaying(page, options)
	if err != nil {
		return nil, err
	}
	s.prewarmMovies(res.Results)
	return s.enrichMovies(ctx, res)
}

func (s *MovieService) GetTopRated(ctx context.Context, page int) (*tmdb.PaginatedResult[MovieDTO], error) {
	res, err := tmdb.GetTopRated(page)
	if err != nil {
		return nil, err
	}
	s.prewarmMovies(res.Results)
	return s.enrichMovies(ctx, res)
}

func (s *MovieService) GetUpcoming(ctx context.Context, withGenres string, page int, options tmdb.RequestOptions) (*tmdb.PaginatedResult[MovieDTO], error) {
	res, err := tmdb.DiscoverUpcoming(withGenres, page, options)
	if err != nil {
		return nil, err
	}
	return s.enrichMovies(ctx, res)
}

func (s *MovieService) SearchMovies(ctx context.Context, query string, page int, language string) (*tmdb.PaginatedResult[MovieDTO], error) {
	res, err := tmdb.SearchMovies(query, page, language)
	if err != nil {
		return nil, err
	}
	return s.enrichMovies(ctx, res)
}

func (s *MovieService) GetMovieByID(ctx context.Context, id int) (*MovieDetailDTO, error) {
	detail, err := tmdb.GetMovieByID(id)
	if err != nil {
		return nil, err
	}

	stats, err := s.ratingRepo.GetMediaRating(ctx, shared.MediaIdentity{ID: id, Type: shared.MediaTypeMovie})
	if err != nil {
		return nil, err
	}

	ratings := NewMediaRatingsResponse(stats, detail.VoteAverage, detail.VoteCount)
	if s.external.Enabled() {
		ratings = ratings.WithExternal(s.external.Get(ctx, string(shared.MediaTypeMovie), id, detail.IMDbID))
	}
	return &MovieDetailDTO{MovieDetail: *detail, Ratings: ratings}, nil
}

func (s *MovieService) enrichMovies(ctx context.Context, res *tmdb.PaginatedResult[tmdb.Movie]) (*tmdb.PaginatedResult[MovieDTO], error) {
	if len(res.Results) == 0 {
		return &tmdb.PaginatedResult[MovieDTO]{
			Page:         res.Page,
			Results:      []MovieDTO{},
			TotalPages:   res.TotalPages,
			TotalResults: res.TotalResults,
		}, nil
	}

	ids := make([]shared.MediaIdentity, len(res.Results))
	for i, m := range res.Results {
		ids[i] = shared.MediaIdentity{ID: m.ID, Type: shared.MediaTypeMovie}
	}

	ratings, err := s.ratingRepo.GetBatchMediaRatings(ctx, ids)
	if err != nil {
		return nil, err
	}

	dtos := make([]MovieDTO, len(res.Results))
	for i, m := range res.Results {
		key := shared.MediaKey{ID: m.ID, Type: shared.MediaTypeMovie}
		dtos[i] = MovieDTO{
			Movie:   m,
			Ratings: NewMediaRatingsResponse(ratings[key], m.VoteAverage, m.VoteCount),
		}
	}

	return &tmdb.PaginatedResult[MovieDTO]{
		Page:         res.Page,
		Results:      dtos,
		TotalPages:   res.TotalPages,
		TotalResults: res.TotalResults,
	}, nil
}

// batchRatings loads popcorn stats for many titles with a single query.
func (s *MovieService) batchRatings(ctx context.Context, t shared.MediaType, ids []int) (map[shared.MediaKey]shared.RatingStats, error) {
	seen := make(map[int]struct{}, len(ids))
	identities := make([]shared.MediaIdentity, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		identities = append(identities, shared.MediaIdentity{ID: id, Type: t})
	}
	return s.ratingRepo.GetBatchMediaRatings(ctx, identities)
}

func movieRatings(m map[shared.MediaKey]shared.RatingStats, mv tmdb.Movie) MediaRatingsResponse {
	return NewMediaRatingsResponse(m[shared.MediaKey{ID: mv.ID, Type: shared.MediaTypeMovie}], mv.VoteAverage, mv.VoteCount)
}

func tvRatings(m map[shared.MediaKey]shared.RatingStats, tv tmdb.TVSeries) MediaRatingsResponse {
	return NewMediaRatingsResponse(m[shared.MediaKey{ID: tv.ID, Type: shared.MediaTypeTV}], tv.VoteAverage, tv.VoteCount)
}

func (s *MovieService) enrichSeries(ctx context.Context, res *tmdb.PaginatedResult[tmdb.TVSeries]) (*tmdb.PaginatedResult[TVSeriesDTO], error) {
	out := &tmdb.PaginatedResult[TVSeriesDTO]{
		Page: res.Page, Results: []TVSeriesDTO{}, TotalPages: res.TotalPages, TotalResults: res.TotalResults,
	}
	if len(res.Results) == 0 {
		return out, nil
	}
	ids := make([]int, len(res.Results))
	for i, t := range res.Results {
		ids[i] = t.ID
	}
	ratings, err := s.batchRatings(ctx, shared.MediaTypeTV, ids)
	if err != nil {
		return nil, err
	}
	for _, t := range res.Results {
		out.Results = append(out.Results, TVSeriesDTO{TVSeries: t, Ratings: tvRatings(ratings, t)})
	}
	return out, nil
}

func (s *MovieService) GetPopularSeries(ctx context.Context, page int) (*tmdb.PaginatedResult[TVSeriesDTO], error) {
	res, err := tmdb.GetPopularSeries(page)
	if err != nil {
		return nil, err
	}
	s.prewarmSeries(res.Results)
	return s.enrichSeries(ctx, res)
}

func (s *MovieService) GetNowAiringSeries(ctx context.Context, page int) (*tmdb.PaginatedResult[TVSeriesDTO], error) {
	res, err := tmdb.GetNowAiringSeries(page)
	if err != nil {
		return nil, err
	}
	s.prewarmSeries(res.Results)
	return s.enrichSeries(ctx, res)
}

func (s *MovieService) GetTopRatedSeries(ctx context.Context, page int) (*tmdb.PaginatedResult[TVSeriesDTO], error) {
	res, err := tmdb.GetTopRatedSeries(page)
	if err != nil {
		return nil, err
	}
	s.prewarmSeries(res.Results)
	return s.enrichSeries(ctx, res)
}

func (s *MovieService) DiscoverSeries(ctx context.Context, genres, tab string, page int) (*tmdb.PaginatedResult[TVSeriesDTO], error) {
	res, err := tmdb.DiscoverSeries(genres, tab, page)
	if err != nil {
		return nil, err
	}
	s.prewarmSeries(res.Results)
	return s.enrichSeries(ctx, res)
}

func (s *MovieService) SearchSeries(ctx context.Context, query string, page int, language string) (*tmdb.PaginatedResult[TVSeriesDTO], error) {
	res, err := tmdb.SearchSeries(query, page, language)
	if err != nil {
		return nil, err
	}
	return s.enrichSeries(ctx, res)
}

func (s *MovieService) GetSimilarSeries(ctx context.Context, id, page int) (*tmdb.PaginatedResult[TVSeriesDTO], error) {
	res, err := tmdb.GetSimilarSeries(id, page)
	if err != nil {
		return nil, err
	}
	return s.enrichSeries(ctx, res)
}

func (s *MovieService) GetSimilar(ctx context.Context, id, page int) (*tmdb.PaginatedResult[MovieDTO], error) {
	res, err := tmdb.GetSimilar(id, page)
	if err != nil {
		return nil, err
	}
	return s.enrichMovies(ctx, res)
}

func (s *MovieService) DiscoverMovies(ctx context.Context, genres, tab string, page int) (*tmdb.PaginatedResult[MovieDTO], error) {
	res, err := tmdb.DiscoverMovies(genres, tab, page)
	if err != nil {
		return nil, err
	}
	s.prewarmMovies(res.Results)
	return s.enrichMovies(ctx, res)
}

// GetSeriesByID returns TV detail with popcorn, tmdb, imdb and rotten_tomatoes ratings.
func (s *MovieService) GetSeriesByID(ctx context.Context, id int) (*TVSeriesDetailDTO, error) {
	detail, err := tmdb.GetSeriesByID(id)
	if err != nil {
		return nil, err
	}
	stats, err := s.ratingRepo.GetMediaRating(ctx, shared.MediaIdentity{ID: id, Type: shared.MediaTypeTV})
	if err != nil {
		return nil, err
	}
	ratings := NewMediaRatingsResponse(stats, detail.VoteAverage, detail.VoteCount)
	if s.external.Enabled() {
		ratings = ratings.WithExternal(s.external.Get(ctx, string(shared.MediaTypeTV), id, ""))
	}
	return &TVSeriesDetailDTO{TVSeriesDetail: *detail, Ratings: ratings}, nil
}

// SeriesIMDbResolver looks up a title's IMDb id through TMDB external_ids
// (TV on detail pages, movies and TV when prewarming list results).
func SeriesIMDbResolver(_ context.Context, mediaType string, tmdbID int) (string, error) {
	switch mediaType {
	case string(shared.MediaTypeTV):
		return tmdb.GetSeriesIMDbID(tmdbID)
	case string(shared.MediaTypeMovie):
		return tmdb.GetMovieIMDbID(tmdbID)
	}
	return "", fmt.Errorf("no resolver for media type %q", mediaType)
}

func (s *MovieService) GetPersonMovieCredits(ctx context.Context, personID int) (*PersonMovieCreditsDTO, error) {
	res, err := tmdb.GetPersonMovieCredits(personID)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(res.Cast)+len(res.Crew))
	for _, c := range res.Cast {
		ids = append(ids, c.ID)
	}
	for _, c := range res.Crew {
		ids = append(ids, c.ID)
	}
	ratings, err := s.batchRatings(ctx, shared.MediaTypeMovie, ids)
	if err != nil {
		return nil, err
	}
	out := &PersonMovieCreditsDTO{Cast: []PersonMovieCreditItemDTO{}, Crew: []MovieDTO{}}
	for _, c := range res.Cast {
		out.Cast = append(out.Cast, PersonMovieCreditItemDTO{PersonMovieCreditItem: c, Ratings: movieRatings(ratings, c.Movie)})
	}
	for _, c := range res.Crew {
		out.Crew = append(out.Crew, MovieDTO{Movie: c, Ratings: movieRatings(ratings, c)})
	}
	return out, nil
}

func (s *MovieService) GetPersonTVCredits(ctx context.Context, personID int) (*PersonTVCreditsDTO, error) {
	res, err := tmdb.GetPersonTVCredits(personID)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(res.Cast)+len(res.Crew))
	for _, c := range res.Cast {
		ids = append(ids, c.ID)
	}
	for _, c := range res.Crew {
		ids = append(ids, c.ID)
	}
	ratings, err := s.batchRatings(ctx, shared.MediaTypeTV, ids)
	if err != nil {
		return nil, err
	}
	out := &PersonTVCreditsDTO{Cast: []PersonTVCreditItemDTO{}, Crew: []TVSeriesDTO{}}
	for _, c := range res.Cast {
		out.Cast = append(out.Cast, PersonTVCreditItemDTO{PersonTVCreditItem: c, Ratings: tvRatings(ratings, c.TVSeries)})
	}
	for _, c := range res.Crew {
		out.Crew = append(out.Crew, TVSeriesDTO{TVSeries: c, Ratings: tvRatings(ratings, c)})
	}
	return out, nil
}

// EnrichMovies attaches ratings.popcorn/tmdb to a raw TMDB movie page.
func (s *MovieService) EnrichMovies(ctx context.Context, res *tmdb.PaginatedResult[tmdb.Movie]) (*tmdb.PaginatedResult[MovieDTO], error) {
	return s.enrichMovies(ctx, res)
}

// prewarmMovies and prewarmSeries warm the external-ratings cache for a list
// page so the first hover popup already has IMDb / Rotten Tomatoes. Non-blocking.
func (s *MovieService) prewarmMovies(items []tmdb.Movie) {
	ids := make([]int, len(items))
	for i, m := range items {
		ids[i] = m.ID
	}
	s.external.Prewarm(string(shared.MediaTypeMovie), ids)
}

func (s *MovieService) prewarmSeries(items []tmdb.TVSeries) {
	ids := make([]int, len(items))
	for i, t := range items {
		ids[i] = t.ID
	}
	s.external.Prewarm(string(shared.MediaTypeTV), ids)
}
