package tvdb

import (
	"os"
	"testing"
)

const APIKey = "DECE3B6B5464C552"

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
	series, err := tvdb.GetSeriesByID(71663)

	if err != nil {
		t.Error(err)
	}

	if series.SeriesName != "The Simpsons" {
		t.Error("ID lookup for '71663' failed.")
	}
}

// TestGetSeriesByIMDBID tests the GetSeriesByIMDBID function.
func TestGetSeriesByIMDBID(t *testing.T) {
	series, err := tvdb.GetSeriesByIMDBID("tt0096697")

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

// TestSeriesGetDetail tests the Series type's GetDetail function.
func TestSeriesGetDetail(t *testing.T) {
	series, err := tvdb.GetSeriesByID(71663)
	if err != nil {
		t.Error(err)
	}

	if series.Seasons != nil {
		t.Error("series.Seasons should be nil.")
	}

	series, err = tvdb.GetSeriesDetail(series.ID)
	if err != nil {
		t.Error(err)
	}

	if series.Seasons == nil {
		t.Error("series.Seasons should not be nil.")
	}
}

func TestMain(m *testing.M) {
	tvdb = NewTVDB(APIKey)
	os.Exit(m.Run())
}
