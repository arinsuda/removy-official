package review_module

import "testing"

func TestBuildRatingResponse_ZeroFilledAndSumMatchesCount(t *testing.T) {
	t.Run("empty buckets are zero-filled 10 steps", func(t *testing.T) {
		res := buildRatingResponse(603, "movie", nil)
		if res.HasRating || res.ReviewCount != 0 || res.AverageRating != 0 {
			t.Fatalf("empty rating response: %+v", res)
		}
		if len(res.Distribution) != 10 {
			t.Fatalf("len(distribution) = %d, want 10", len(res.Distribution))
		}
		for i, b := range res.Distribution {
			wantRating := float32(i+1) * 0.5
			if b.Rating != wantRating || b.Count != 0 {
				t.Errorf("bucket[%d] = %+v, want {Rating: %v, Count: 0}", i, b, wantRating)
			}
		}
	})

	t.Run("partial buckets keep 10 steps and sum(distribution) == review_count", func(t *testing.T) {
		res := buildRatingResponse(603, "movie", []RatingBucket{
			{Rating: 2.5, Count: 1},
			{Rating: 4.5, Count: 2},
			{Rating: 5.0, Count: 1},
		})
		if !res.HasRating || res.ReviewCount != 4 || res.AverageRating != 4.1 {
			t.Fatalf("rating response = %+v, want count=4 avg=4.1", res)
		}
		if len(res.Distribution) != 10 {
			t.Fatalf("len(distribution) = %d, want 10", len(res.Distribution))
		}
		sum := 0
		for _, b := range res.Distribution {
			sum += b.Count
		}
		if sum != res.ReviewCount {
			t.Fatalf("sum(distribution) = %d, want review_count = %d", sum, res.ReviewCount)
		}
	})
}

func TestBuildUserReviewsSummary(t *testing.T) {
	empty := buildUserReviewsSummary(userReviewSummaryRow{})
	if empty.Count != 0 || empty.AverageRating != 0 || empty.FilterCounts != (ReviewFilterCounts{}) {
		t.Fatalf("empty summary = %+v", empty)
	}
	got := buildUserReviewsSummary(userReviewSummaryRow{Count: 3, RatingSum: 12.5, High: 2, Low: 1, WithText: 2})
	want := UserReviewsSummary{Count: 3, AverageRating: 4.2, FilterCounts: ReviewFilterCounts{All: 3, High: 2, Low: 1, WithText: 2}}
	if got != want {
		t.Fatalf("summary = %+v, want %+v", got, want)
	}
}

func TestClampPage(t *testing.T) {
	for in, want := range map[int]int{-5: 1, 0: 1, 1: 1, MaxReviewsPage: MaxReviewsPage, MaxReviewsPage + 1: MaxReviewsPage, int(^uint(0) >> 1): MaxReviewsPage} {
		if got := clampPage(in); got != want {
			t.Errorf("clampPage(%d) = %d, want %d", in, got, want)
		}
	}
}
