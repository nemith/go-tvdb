package tvdb

import (
	"os"
	"testing"
	"time"
)

const (
	apiKey     = "DECE3B6B5464C552"
	testUser   = "34A8615ABE815874"
	simpsonsID = 71663
	futuramaID = 73871
)

var tvdb *TVDB

// TestGetSeries tests the GetSeries function.
func TestGetSeries(t *testing.T) {
	seriesList, err := tvdb.GetSeries("The Simpsons")

	if err != nil {
		t.Error(err)
	}

	for _, series := range seriesList {
		if series.SeriesName == "The Simpsons" {
			return
		}
	}

	t.Error("No 'The Simpsons' title could be found.")
}

// TestGetSeriesByID tests the GetSeriesByID function.
func TestGetSeriesByID(t *testing.T) {
	series, err := tvdb.GetSeriesByID(simpsonsID)

	if err != nil {
		t.Error(err)
	}

	if series.SeriesName != "The Simpsons" {
		t.Error("ID lookup for 'simpsonsID' failed.")
	}
}

// TestGetSeriesByRemoteID tests the GetSeriesByRemoteID function.
func TestGetSeriesByRemoteID(t *testing.T) {
	series, err := tvdb.GetSeriesByRemoteID(IMDB, "tt0096697")

	if err != nil {
		t.Error(err)
	}

	if series.SeriesName != "The Simpsons" {
		t.Error("IMDb ID lookup for 'tt0096697' failed.")
	}
}

// TestSearchSeries tests the SearchSeries function.
func TestSearchSeries(t *testing.T) {
	seriesList, err := tvdb.SearchSeries("The Simpsons", 5)

	if err != nil {
		t.Error(err)
	}

	for _, series := range seriesList {
		if series.SeriesName == "The Simpsons" {
			return
		}
	}

	t.Error("No 'The Simpsons' title could be found.")
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
	series, err := tvdb.GetSeriesFull(simpsonsID)
	if err != nil {
		t.Error(err)
	}

	if series.Seasons == nil {
		t.Error("series.Seasons should not be nil.")
	}
}

func TestUserFav(t *testing.T) {
	// Test user with one favorite
	favs, err := tvdb.UserFav("34A8615ABE815874")
	if err != nil {
		t.Error(err)
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

func TestMain(m *testing.M) {
	tvdb = NewTVDB(apiKey)
	os.Exit(m.Run())
}
