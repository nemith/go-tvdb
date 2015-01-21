package tvdb

import (
	"os"
	"testing"
	"time"
)

const (
	apiKey       = "90D7DF3AE9E4841E"
	testUser     = "34A8615ABE815874"
	simpsonsID   = 71663
	simpsonsIMDB = "tt0096697"
	simpsonsName = "The Simpsons"
	futuramaID   = 73871
)

var tvdb *TVDB

// TestGetSeries tests the GetSeries function.
func TestGetSeries(t *testing.T) {
	t.Logf("Finding series with name '%s'", simpsonsName)
	seriesList, err := tvdb.GetSeries(simpsonsName)

	if err != nil {
		t.Fatal(err)
	}

	for _, series := range seriesList {
		if series.Name == simpsonsName {
			return
		}
	}

	t.Errorf("Expected to find series '%s' got '%s'", simpsonsName, seriesList)
}

// TestGetSeriesByID tests the GetSeriesByID function.
func TestGetSeriesByID(t *testing.T) {
	t.Logf("Getting series with id '%d'", simpsonsID)
	series, err := tvdb.GetSeriesByID(simpsonsID)
	if err != nil {
		t.Fatal(err)
	}

	if series.Name != simpsonsName {
		t.Errorf("Lookup for ID '%d'. Expected name of '%s' got '%s'",
			simpsonsID, simpsonsName, series.Name)
	}
}

// TestGetSeriesByRemoteID tests the GetSeriesByRemoteID function.
func TestGetSeriesByRemoteID(t *testing.T) {
	t.Logf("Getting series with IMDB ID '%s'", simpsonsIMDB)
	series, err := tvdb.GetSeriesByRemoteID(IMDB, simpsonsIMDB)
	if err != nil {
		t.Fatal(err)
	}

	if series.Name != simpsonsName {
		t.Errorf("Expectted series name of '%s' got '%s' for IMDB ID of '%s' failed.")
	}
}

// TestSearchSeries tests the SearchSeries function.
func TestSearchSeries(t *testing.T) {
	t.Logf("Searching for series with name '%s'", simpsonsName)
	seriesList, err := tvdb.SearchSeries(simpsonsName, 5)
	if err != nil {
		t.Fatal(err)
	}

	for _, series := range seriesList {
		if series.Name == simpsonsName {
			return
		}
	}

	t.Errorf("Expected to find series '%s' got '%s'", simpsonsName, seriesList)
}

func seriesIDExists(favs []int, seriesID int) bool {
	for _, fav := range favs {
		if fav == seriesID {
			return true
		}
	}
	return false
}

// TestSeriesGetDetail tests the Series type's GetDetail function.
func TestGetSeriesFull(t *testing.T) {
	t.Logf("Getting Full series for seriesID '%d'", simpsonsID)
	series, err := tvdb.GetSeriesFull(simpsonsID)
	if err != nil {
		t.Fatal(err)
	}

	if series.Seasons == nil {
		t.Error("series.Seasons should not be nil.")
	}
}

func TestUserFav(t *testing.T) {
	t.Logf("Querying favorites for userID '%s'", testUser)
	// Test user with one favorite
	favs, err := tvdb.UserFav(testUser)
	if err != nil {
		t.Fatal(err)
	}

	if !seriesIDExists(favs, simpsonsID) {
		t.Errorf("Expected to find seriesID '%d' got %s", simpsonsID, favs)
	}
}

func TestUserFavAddRemove(t *testing.T) {
	t.Logf("Adding series '%d to user '%s' favorites", futuramaID, testUser)
	favs, err := tvdb.UserFavAdd(testUser, futuramaID)
	if err != nil {
		t.Fatal(err)
	}

	if !seriesIDExists(favs, futuramaID) {
		t.Errorf("Expected to find seriesID '%d' got %s", futuramaID, favs)
	}
	time.Sleep(1 * time.Second)
	t.Logf("Removing series '%d' from user '%s' favorites", futuramaID, testUser)
	favs, err = tvdb.UserFavRemove(testUser, futuramaID)
	if err != nil {
		t.Fatal(err)
	}
	if seriesIDExists(favs, futuramaID) {
		t.Errorf("Expected to NOT find seriesID '%d got %s", futuramaID, favs)
	}
}

func TestGetRatingsForUser(t *testing.T) {
	t.Logf("Getting ratings for user '%s'", testUser)
	ratings, err := tvdb.GetRatingsForUser(testUser)
	if err != nil {
		t.Fatal(err)
	}

	if len(ratings) < 1 {
		t.Errorf("Expected at least one rating")
	} else {
		rating := ratings[0]

		if rating.ID <= 0 {
			t.Errorf("Expected non-zero seriesID")
		}

		if rating.CommunityRating <= 0 {
			t.Errorf("Expected a non-zero Community rating")
		}
	}

}

func TestSetUserRatingSeries(t *testing.T) {
	rating := 7
	t.Logf("Setting rating for user '%s' and for series id '%d' to '%d'", testUser, simpsonsID, rating)
	if err := tvdb.SetUserRatingSeries(testUser, simpsonsID, rating); err != nil {
		t.Fatal(err)
	}
}

func TestUserLang(t *testing.T) {
	t.Logf("Getting prefered language for user '%s'", testUser)
	lang, err := tvdb.UserLang(testUser)
	if err != nil {
		t.Fatal(err)
	}

	if lang.Abbr != "en" {
		t.Errorf("Expected language abbr of '%s' got '%s'", "en", lang.Abbr)
	}

	if lang.Name != "English" {
		t.Errorf("Expected language name of '%s' got '%s'", "English", lang.Name)
	}
}

func TestMain(m *testing.M) {
	tvdb = NewTVDB(apiKey)
	os.Exit(m.Run())
}
