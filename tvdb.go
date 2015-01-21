package tvdb

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// PipeList type representing pipe-separated string values.
type PipeList []string

// UnmarshalXML unmarshals an XML element with string value into a pip-separated list of strings.
func (pipeList *PipeList) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	content := ""
	if err := decoder.DecodeElement(&content, &start); err != nil {
		return err
	}

	*pipeList = strings.Split(strings.Trim(content, "|"), "|")
	return nil
}

// Episode represents a TV show episode on TheTVDB.
type Episode struct {
	ID                    uint64   `xml:"id"`
	CombinedEpisodeNumber string   `xml:"Combined_episodenumber"`
	CombinedSeason        uint64   `xml:"Combined_season"`
	DvdChapter            string   `xml:"DVD_chapter"`
	DvdDiscID             string   `xml:"DVD_discid"`
	DvdEpisodeNumber      string   `xml:"DVD_episodenumber"`
	DvdSeason             string   `xml:"DVD_season"`
	Director              PipeList `xml:"Director"`
	EpImgFlag             string   `xml:"EpImgFlag"`
	EpisodeName           string   `xml:"EpisodeName"`
	EpisodeNumber         uint64   `xml:"EpisodeNumber"`
	FirstAired            string   `xml:"FirstAired"`
	GuestStars            string   `xml:"GuestStars"`
	ImdbID                string   `xml:"IMDB_ID"`
	Language              string   `xml:"Language"`
	Overview              string   `xml:"Overview"`
	ProductionCode        string   `xml:"ProductionCode"`
	Rating                string   `xml:"Rating"`
	RatingCount           string   `xml:"RatingCount"`
	SeasonNumber          uint64   `xml:"SeasonNumber"`
	Writer                PipeList `xml:"Writer"`
	AbsoluteNumber        string   `xml:"absolute_number"`
	Filename              string   `xml:"filename"`
	LastUpdated           string   `xml:"lastupdated"`
	SeasonID              uint64   `xml:"seasonid"`
	SeriesID              uint64   `xml:"seriesid"`
	ThumbAdded            string   `xml:"thumb_added"`
	ThumbHeight           string   `xml:"thumb_height"`
	ThumbWidth            string   `xml:"thumb_width"`
}

// Series represents TV show on TheTVDB.
type Series struct {
	ID            uint64   `xml:"id"`
	Actors        PipeList `xml:"Actors"`
	AirsDayOfWeek string   `xml:"Airs_DayOfWeek"`
	AirsTime      string   `xml:"Airs_Time"`
	ContentRating string   `xml:"ContentRating"`
	FirstAired    string   `xml:"FirstAired"`
	Genre         PipeList `xml:"Genre"`
	ImdbID        string   `xml:"IMDB_ID"`
	Language      string   `xml:"Language"`
	Network       string   `xml:"Network"`
	NetworkID     string   `xml:"NetworkID"`
	Overview      string   `xml:"Overview"`
	Rating        string   `xml:"Rating"`
	RatingCount   string   `xml:"RatingCount"`
	Runtime       string   `xml:"Runtime"`
	SeriesID      string   `xml:"SeriesID"`
	SeriesName    string   `xml:"SeriesName"`
	Status        string   `xml:"Status"`
	Added         string   `xml:"added"`
	AddedBy       string   `xml:"addedBy"`
	Banner        string   `xml:"banner"`
	Fanart        string   `xml:"fanart"`
	LastUpdated   string   `xml:"lastupdated"`
	Poster        string   `xml:"poster"`
	Zap2ItID      string   `xml:"zap2it_id"`
	Seasons       map[uint64][]*Episode
}

type RemoteService string

const (
	IMDB   = RemoteService("imdbid")
	Zap2it = RemoteService("zap2it")
)

// data is the response back from the server
type Data struct {
	Series   []*Series  `xml:"Series,omitempty"`
	Episodes []*Episode `xml:"Episode,omitempty"`
}

type TVDB struct {
	APIKey string
	//defaultLang string
}

func NewTVDB(apiKey string) *TVDB {
	return &TVDB{
		APIKey: apiKey,
	}
}

func getResponse(url string) (*Data, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := &Data{}
	d := xml.NewDecoder(resp.Body)
	if err = d.Decode(data); err != nil {
		return nil, err
	}

	return data, nil
}

func (t *TVDB) baseURL() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   "thetvdb.com",
	}
}

func (t *TVDB) apiURL(path string, query url.Values) *url.URL {
	url := t.baseURL()
	url.Path = fmt.Sprintf("api/%s", path)
	url.RawQuery = query.Encode()
	return url
}

func (t *TVDB) staticAPIURL(path string) *url.URL {
	url := t.baseURL()
	url.Path = fmt.Sprintf("api/%s/%s", t.APIKey, path)
	return url
}

// GetSeries gets a list of TV series by name, by performing a simple search.
func (t *TVDB) GetSeries(name string) ([]*Series, error) {
	u := t.apiURL("GetSeries.php", url.Values{
		"seriesname": []string{name},
	})
	data, err := getResponse(u.String())
	if err != nil {
		return nil, err
	}
	return data.Series, nil
}

// GetSeriesByID gets a TV series by ID.
func (t *TVDB) GetSeriesByID(id uint64) (*Series, error) {
	u := t.staticAPIURL(fmt.Sprintf("series/%d/en.xml", id))
	data, err := getResponse(u.String())
	if err != nil {
		return nil, err
	}

	if len(data.Series) != 1 {
		return nil, fmt.Errorf("Got too many series (expected: 1, got: %d)", len(data.Series))
	}

	return data.Series[0], nil
}

// GetSeriesByIMDBID gets series from IMDb's ID.
func (t *TVDB) GetSeriesByRemoteID(service RemoteService, id string) (*Series, error) {
	query := url.Values{}
	query.Set(string(service), id)
	u := t.apiURL("GetSeriesByRemoteID.php", query)
	data, err := getResponse(u.String())
	if err != nil {
		return nil, err
	}

	if len(data.Series) != 1 {
		return nil, fmt.Errorf("Got too many series (expected: 1, got: %d)", len(data.Series))
	}

	return data.Series[0], nil
}

// GetDetail gets more detail for a TV show, including information on it's episodes.
func (t *TVDB) GetSeriesDetail(seriesID uint64) (*Series, error) {
	u := t.staticAPIURL(fmt.Sprintf("series/%d/all/en.xml", seriesID))
	data, err := getResponse(u.String())
	if err != nil {
		return nil, err
	}

	if len(data.Series) != 1 {
		return nil, fmt.Errorf("Got too many series (expected: 1, got: %d)", len(data.Series))
	}

	series := data.Series[0]
	if series.Seasons == nil {
		series.Seasons = make(map[uint64][]*Episode, len(data.Episodes))
	}

	for _, episode := range data.Episodes {
		series.Seasons[episode.SeasonNumber] = append(series.Seasons[episode.SeasonNumber], episode)
	}
	return series, nil
}

var reSearchSeries = regexp.MustCompile(`(?P<before><a href="/\?tab=series&amp;id=)(?P<seriesId>\d+)(?P<after>\&amp;lid=\d*">)`)

// SearchSeries searches for TV shows by name, using the more sophisticated
// search on TheTVDB's homepage. This is the recommended search method.
func (t *TVDB) SearchSeries(name string, maxResults int) ([]Series, error) {
	u := t.baseURL()
	query := url.Values{
		"string":         []string{name},
		"searchseriesid": []string{""},
		"tab":            []string{"listseries"},
		"function":       []string{"Search"},
	}
	u.RawQuery = query.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	results := reSearchSeries.FindAllStringSubmatch(buf.String(), -1)

	if len(results) < maxResults {
		maxResults = len(results)
	}
	seriesList := make([]Series, maxResults)

	for _, result := range results {
		seriesID := uint64(0)
		var series *Series
		seriesID, err = strconv.ParseUint(string(result[2]), 10, 64)
		if err != nil {
			continue
		}

		series, err = t.GetSeriesByID(seriesID)
		if err != nil {
			// Some series can't be found, so we will ignore these.
			if _, ok := err.(*xml.SyntaxError); ok {
				err = nil
				continue
			} else {
				return seriesList, err
			}
		}

		seriesList = append(seriesList, *series)

		if len(seriesList) == maxResults {
			break
		}
	}
	return seriesList, nil
}
